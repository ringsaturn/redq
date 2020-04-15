package redq

type QueuedMessage []byte

func (qm QueuedMessage) Message() []byte {
	return []byte(qm)
}
