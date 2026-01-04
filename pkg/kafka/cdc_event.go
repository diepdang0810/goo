package kafka

type CDCEvent struct {
	Op string `json:"__op"`
}

func (e *CDCEvent) IsCreated() bool {
	return e.Op == "c"
}

func (e *CDCEvent) IsUpdated() bool {
	return e.Op == "u"
}

func (e *CDCEvent) IsDeleted() bool {
	return e.Op == "d"
}

func (e *CDCEvent) IsSnapshot() bool {
	return e.Op == "r"
}
