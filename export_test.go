package jobq

type Worker worker

func (w *Worker) ID() int {
	return w.id
}

func (w *Worker) kerker() chan bool {
	return w.stopC
}
