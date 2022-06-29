package helper

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
)

type Map map[string]interface{}

func DBErrResponse(err error) Map {
	m := Map{}

	if mongo.IsDuplicateKeyError(err) {
		m["error"] = "oops, this alias has been used before. Choose another one" 
		return m
	}

	switch {
	case err == mongo.ErrNoDocuments:
		m["error"] = "Data not found"
		return m
	
	case err == mongo.ErrNilDocument:
		m["error"] = "Data not found"
		return m

	case err == mongo.ErrNilValue:
		m["error"] = "Data not found"
		return m

	default:
		m["error"] = "Unknown error occured. Please try again in a few seconds" 
		return m 
	}
}

func InternalError() string {
	return "Internal server error. Please try again in a few seconds"
}

func WriteErr(err error) Map {
	mapErr := Map{}

	if castedObject, ok := err.(validator.ValidationErrors); ok {
		for _, err := range castedObject {
			switch err.Tag() {
			case "required":
               mapErr["error"] = fmt.Sprintf("%s is required", err.Field())
			   return mapErr

			case "alphanum":
				mapErr["error"] = fmt.Sprintf("%s must only contain alphanumeric", err.Field())
				return mapErr
			
			case "min":
                mapErr["error"] = fmt.Sprintf("%s value must be greater than %s", err.Field(), err.Param())
				return mapErr
			
			case "max":
                mapErr["error"] = fmt.Sprintf("%s value must be lower than %s", err.Field(), err.Param())
				return mapErr

			case "oneof":
				mapErr["error"] = "Unknown value. Please try again"
				return mapErr
			}
		}
	}

	return Map{
		"error": err.Error(),
	}
}