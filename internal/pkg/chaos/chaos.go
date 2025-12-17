package chaos

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"kube-monkey/internal/pkg/config"
	"kube-monkey/internal/pkg/kubernetes"
	"kube-monkey/internal/pkg/victims"
)

type Chaos struct {
	killAt time.Time
	victim victims.Victim
}

// New creates a new Chaos instance
func New(killtime time.Time, victim victims.Victim) *Chaos {
	// TargetPodName will be populated at time of termination
	return &Chaos{
		killAt: killtime,
		victim: victim,
	}
}

func (c *Chaos) Victim() victims.Victim {
	return c.victim
}

func (c *Chaos) KillAt() time.Time {
	return c.killAt
}

// Schedule the execution of Chaos
func (c *Chaos) Schedule(resultchan chan<- *Result) {
	time.Sleep(c.DurationToKillTime())
	c.Execute(resultchan)
}

// DurationToKillTime calculates the duration from now until Chaos.killAt
func (c *Chaos) DurationToKillTime() time.Duration {
	return time.Until(c.killAt)
}

// Execute exposed function that calls the actual execution of the chaos, i.e. termination of pods
// The result is sent back over the channel provided
func (c *Chaos) Execute(resultchan chan<- *Result) {
	// Create kubernetes clientset
	clientset, dynamicClient, err := kubernetes.CreateClient()
	if err != nil {
		resultchan <- c.NewResult(err)
		return
	}

	victimClient := victims.NewVictimClient(clientset, dynamicClient)

	err = c.verifyExecution(victimClient)
	if err != nil {
		resultchan <- c.NewResult(err)
		return
	}

	err = c.terminate(victimClient)
	if err != nil {
		resultchan <- c.NewResult(err)
		return
	}

	// Send a success msg
	resultchan <- c.NewResult(nil)
}

// Verify if the victim has opted out since scheduling
func (c *Chaos) verifyExecution(client victims.VictimKubeClient) error {
	// Is victim still enrolled in kube-monkey
	enrolled, err := c.Victim().IsEnrolled(client)
	if err != nil {
		return err
	}

	if !enrolled {
		return fmt.Errorf("%s %s is no longer enrolled in kube-monkey. Skipping", c.Victim().Kind(), c.Victim().Name())
	}

	// Has the victim been blacklisted since scheduling?
	if c.Victim().IsBlacklisted() {
		return fmt.Errorf("%s %s is blacklisted. Skipping", c.Victim().Kind(), c.Victim().Name())
	}

	// Has the victim been removed from the whitelist since scheduling?
	if !c.Victim().IsWhitelisted() {
		return fmt.Errorf("%s %s is not whitelisted. Skipping", c.Victim().Kind(), c.Victim().Name())
	}

	// Send back valid for termination
	return nil
}

// The termination type and value is processed here
func (c *Chaos) terminate(client victims.VictimKubeClient) error {
	killType, err := c.Victim().KillType(client)
	if err != nil {
		return errors.Wrapf(err, "Failed to check KillType label for %s %s", c.Victim().Kind(), c.Victim().Name())
	}

	killValue, err := c.getKillValue(client)

	// KillAll is the only kill type that does not require a kill-value
	if killType != config.KillAllLabelValue && err != nil {
		return err
	}

	// Validate killtype
	switch killType {
	case config.KillFixedLabelValue:
		return c.Victim().DeleteRandomPods(client, killValue)
	case config.KillAllLabelValue:
		killNum, err := c.Victim().KillNumberForKillingAll(client)
		if err != nil {
			return err
		}
		return c.Victim().DeleteRandomPods(client, killNum)
	case config.KillRandomMaxLabelValue:
		killNum, err := c.Victim().KillNumberForMaxPercentage(client, killValue)
		if err != nil {
			return err
		}
		return c.Victim().DeleteRandomPods(client, killNum)
	case config.KillFixedPercentageLabelValue:
		killNum, err := c.Victim().KillNumberForFixedPercentage(client, killValue)
		if err != nil {
			return err
		}
		return c.Victim().DeleteRandomPods(client, killNum)
	default:
		return fmt.Errorf("failed to recognize KillType label for %s %s", c.Victim().Kind(), c.Victim().Name())
	}
}

func (c *Chaos) getKillValue(client victims.VictimKubeClient) (int, error) {
	killValue, err := c.Victim().KillValue(client)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to check KillValue label for %s %s", c.Victim().Kind(), c.Victim().Name())
	}

	return killValue, nil
}

// NewResult creates a ChaosResult instance
func (c *Chaos) NewResult(e error) *Result {
	return &Result{
		chaos: c,
		err:   e,
	}
}
