package base

/*
Response
	Status: {Code: int (200/404),  Message :}
		200 success
		400 other and external error
		403 No permission
		412 Validate error
		500 internal error
	Message : {Code : int, Message:}
	Data:
		validate error : {Errors:['username':'more than 5 word',]}
*/

// ResponseErrorMessage Response message
type ResponseErrorMessage struct {
	Code    string
	Message string
	Input   string
}

// ResponseStatus Response status
type ResponseStatus struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response Response base
type Response[T any] struct {
	*ResponseStatus
	Data     T                 `json:"data"`
	Paginate *ResponsePaginate `json:",omitempty"`
}

type ResponsePaginate struct {
	Page  int64
	Size  int64
	Total int64
}

// ResponseValidateMessage Response calidate message
type ResponseValidateMessage struct {
	Errors map[string]*ResponseErrorMessage
}
