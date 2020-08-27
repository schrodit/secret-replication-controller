package errors

// This packages defines specific error reasons.

const (
	// InternalError defines an internal error in the controller
	InternalError Reason = "InternalError"
	// CreateError defines an error that occurred when a replicated secret could not be created
	CreateError Reason = "CreateError"
	// UpdateError defines an error that occurred when a replicated secret could not be updated
	UpdateError Reason = "UpdateError"
	// InvalidNamespace defines an error reason that is thrown when a namespaces does not exist or cannot be validated
	InvalidNamespace Reason = "InvalidNamespace"
)
