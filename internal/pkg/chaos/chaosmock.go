package chaos

import (
	"time"

	"kube-monkey/internal/pkg/victims"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAMESPACE  = metav1.NamespaceDefault
	IDENTIFIER = "kube-monkey-id"
	KIND       = "Pod"
	NAME       = "name"
)

type VictimMock struct {
	mock.Mock
	*victims.VictimBase
}

func (vm *VictimMock) IsEnrolled(client victims.VictimKubeClient) (bool, error) {
	args := vm.Called(client)
	return args.Bool(0), args.Error(1)
}

func (vm *VictimMock) KillType(client victims.VictimKubeClient) (string, error) {
	args := vm.Called(client)
	return args.String(0), args.Error(1)
}

func (vm *VictimMock) KillValue(client victims.VictimKubeClient) (int, error) {
	args := vm.Called(client)
	return args.Int(0), args.Error(1)
}

func (vm *VictimMock) DeleteRandomPod(client victims.VictimKubeClient) error {
	args := vm.Called(client)
	return args.Error(0)
}

func (vm *VictimMock) DeleteRandomPods(client victims.VictimKubeClient, killValue int) error {
	args := vm.Called(client, killValue)
	return args.Error(0)
}

func (vm *VictimMock) KillNumberForKillingAll(client victims.VictimKubeClient) (int, error) {
	args := vm.Called(client)
	return args.Int(0), args.Error(1)
}

func (vm *VictimMock) KillNumberForMaxPercentage(client victims.VictimKubeClient, killValue int) (int, error) {
	args := vm.Called(client, killValue)
	return args.Int(0), args.Error(1)
}

func (vm *VictimMock) KillNumberForFixedPercentage(client victims.VictimKubeClient, killValue int) (int, error) {
	args := vm.Called(client, killValue)
	return args.Int(0), args.Error(1)
}

func (vm *VictimMock) IsBlacklisted() bool {
	args := vm.Called()
	return args.Bool(0)
}

func (vm *VictimMock) IsWhitelisted() bool {
	args := vm.Called()
	return args.Bool(0)
}

func NewVictimMock() *VictimMock {
	v := victims.New(KIND, NAME, NAMESPACE, IDENTIFIER, 1)
	return &VictimMock{
		VictimBase: v,
	}
}

func NewMock() *Chaos {
	return &Chaos{
		killAt: time.Now(),
		victim: NewVictimMock(),
	}
}
