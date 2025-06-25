package cluster

type Slot struct {
	Start int32
	End   int32
}

type Node struct {
	ID   string
	Slot Slot
	Host string
	Port string
}
