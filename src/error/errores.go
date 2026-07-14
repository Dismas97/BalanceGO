package error

type ENoEncontrado struct {
	Codigo  int
	Mensaje string
}
func (e *ENoEncontrado) Error() string {
	return "Error: No encontrado"
}
