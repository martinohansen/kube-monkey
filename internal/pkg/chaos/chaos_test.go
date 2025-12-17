package chaos

import (
	"errors"
	"testing"

	"kube-monkey/internal/pkg/config"
	"kube-monkey/internal/pkg/victims"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type ChaosTestSuite struct {
	suite.Suite
	chaos        *Chaos
	client       kube.Interface
	victimClient victims.VictimKubeClient
}

func (s *ChaosTestSuite) SetupTest() {
	s.chaos = NewMock()
	s.client = fake.NewSimpleClientset()
	s.victimClient = victims.NewVictimClient(s.client, nil)
}

func (s *ChaosTestSuite) TestVerifyExecutionNotEnrolled() {
	v := s.chaos.victim.(*VictimMock)
	v.On("IsEnrolled", s.victimClient).Return(false, nil)
	err := s.chaos.verifyExecution(s.victimClient)
	v.AssertExpectations(s.T())
	s.EqualError(err, v.Kind()+" "+v.Name()+" is no longer enrolled in kube-monkey. Skipping")
}

func (s *ChaosTestSuite) TestVerifyExecutionBlacklisted() {
	v := s.chaos.victim.(*VictimMock)
	v.On("IsEnrolled", s.victimClient).Return(true, nil)
	v.On("IsBlacklisted").Return(true)
	err := s.chaos.verifyExecution(s.victimClient)
	v.AssertExpectations(s.T())
	s.EqualError(err, v.Kind()+" "+v.Name()+" is blacklisted. Skipping")
}

func (s *ChaosTestSuite) TestVerifyExecutionNotWhitelisted() {
	v := s.chaos.victim.(*VictimMock)
	v.On("IsEnrolled", s.victimClient).Return(true, nil)
	v.On("IsBlacklisted").Return(false)
	v.On("IsWhitelisted").Return(false)
	err := s.chaos.verifyExecution(s.victimClient)
	v.AssertExpectations(s.T())
	s.EqualError(err, v.Kind()+" "+v.Name()+" is not whitelisted. Skipping")
}

func (s *ChaosTestSuite) TestVerifyExecutionWhitelisted() {
	v := s.chaos.victim.(*VictimMock)
	v.On("IsEnrolled", s.victimClient).Return(true, nil)
	v.On("IsBlacklisted").Return(false)
	v.On("IsWhitelisted").Return(true)
	err := s.chaos.verifyExecution(s.victimClient)
	v.AssertExpectations(s.T())
	s.NoError(err)
}

func (s *ChaosTestSuite) TestTerminateKillTypeError() {
	v := s.chaos.victim.(*VictimMock)
	err := errors.New("KillType Error")
	v.On("KillType", s.victimClient).Return("", err)

	s.NotNil(s.chaos.terminate(s.victimClient))
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestTerminateKillValueError() {
	v := s.chaos.victim.(*VictimMock)
	errMsg := "KillValue Error"
	v.On("KillType", s.victimClient).Return(config.KillFixedLabelValue, nil)
	v.On("KillValue", s.victimClient).Return(0, errors.New(errMsg))
	s.NotNil(s.chaos.terminate(s.victimClient))
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestTerminateKillFixed() {
	v := s.chaos.victim.(*VictimMock)
	killValue := 1
	v.On("KillType", s.victimClient).Return(config.KillFixedLabelValue, nil)
	v.On("KillValue", s.victimClient).Return(killValue, nil)
	v.On("DeleteRandomPods", s.victimClient, killValue).Return(nil)
	_ = s.chaos.terminate(s.victimClient)
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestTerminateAllPods() {
	v := s.chaos.victim.(*VictimMock)
	v.On("KillType", s.victimClient).Return(config.KillAllLabelValue, nil)
	v.On("KillValue", s.victimClient).Return(0, nil)
	v.On("KillNumberForKillingAll", s.victimClient).Return(0, nil)
	v.On("DeleteRandomPods", s.victimClient, 0).Return(nil)
	_ = s.chaos.terminate(s.victimClient)
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestTerminateKillRandomMaxPercentage() {
	v := s.chaos.victim.(*VictimMock)
	killValue := 1
	v.On("KillType", s.victimClient).Return(config.KillRandomMaxLabelValue, nil)
	v.On("KillValue", s.victimClient).Return(killValue, nil)
	v.On("KillNumberForMaxPercentage", s.victimClient, mock.AnythingOfType("int")).Return(0, nil)
	v.On("DeleteRandomPods", s.victimClient, 0).Return(nil)
	_ = s.chaos.terminate(s.victimClient)
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestTerminateKillFixedPercentage() {
	v := s.chaos.victim.(*VictimMock)
	killValue := 1
	v.On("KillType", s.victimClient).Return(config.KillFixedPercentageLabelValue, nil)
	v.On("KillValue", s.victimClient).Return(killValue, nil)
	v.On("KillNumberForFixedPercentage", s.victimClient, mock.AnythingOfType("int")).Return(0, nil)
	v.On("DeleteRandomPods", s.victimClient, 0).Return(nil)
	_ = s.chaos.terminate(s.victimClient)
	v.AssertExpectations(s.T())
}

func (s *ChaosTestSuite) TestInvalidKillType() {
	v := s.chaos.victim.(*VictimMock)
	v.On("KillType", s.victimClient).Return("InvalidKillTypeHere", nil)
	v.On("KillValue", s.victimClient).Return(0, nil)
	err := s.chaos.terminate(s.victimClient)
	v.AssertExpectations(s.T())
	s.NotNil(err)
}

func (s *ChaosTestSuite) TestGetKillValue() {
	v := s.chaos.victim.(*VictimMock)
	killValue := 5
	v.On("KillValue", s.victimClient).Return(killValue, nil)
	result, err := s.chaos.getKillValue(s.victimClient)
	s.Nil(err)
	s.Equal(killValue, result)
}

func (s *ChaosTestSuite) TestGetKillValueReturnsError() {
	v := s.chaos.victim.(*VictimMock)
	v.On("KillValue", s.victimClient).Return(0, errors.New("InvalidKillValue"))
	_, err := s.chaos.getKillValue(s.victimClient)
	s.NotNil(err)
}

// Disabling test
// See https://github.com/asobti/kube-monkey/issues/126
//func (s *ChaosTestSuite) TestDurationToKillTime() {
//	t := s.chaos.DurationToKillTime()
//	s.WithinDuration(s.chaos.KillAt(), time.Now(), t+time.Millisecond)
//}

func TestSuite(t *testing.T) {
	suite.Run(t, new(ChaosTestSuite))
}
