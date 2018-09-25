package ae

import (
	"errors"

	"context"

	"google.golang.org/appengine/log"
)

func LogError(c context.Context, msg string, err error) error {
	log.Errorf(c, "%s: %s", msg, err.Error())
	return errors.New(msg)
}
