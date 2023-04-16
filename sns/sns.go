package sns

// SNS is testable interface.
type SNS interface {
	Post(post string) error
	String() string
}
