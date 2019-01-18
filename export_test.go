package jobq

type Worker worker

func (w *Worker) ID() int {
	return w.id
}
