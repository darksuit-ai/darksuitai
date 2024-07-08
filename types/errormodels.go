package types


type Error struct { 
	ResponseCode      int    `json:"response code"` 
	Message           string `json:"message"` 
	Detail            string `json:"detail"` 
	ExternalReference string `json:"ext_ref"` 
 }