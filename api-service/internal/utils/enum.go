package utils

// new type: TaskType
type TaskType int

// four task types
const (
	claimWithdraw TaskType = iota + 1 // start from 1
	claimDeposit
	approval
	unapproval
)

var taskTypeStr = []string{"withdraw", "deposit", "approval", "unapproval"}

// return string
func (t TaskType) String() string {
	return taskTypeStr[t-1]
}

// return index(int)
func (t TaskType) Index() int {
	return int(t)
}
