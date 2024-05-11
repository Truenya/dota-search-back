package condlog

import log "github.com/sirupsen/logrus"

func TraceCond(err error) {
	if err != nil {
		log.Traceln(err)
	}
}

func WarnCond(msg string, err error) {
	if err != nil {
		log.Warnf(msg, err)
	}
}

func ErrCond(msg string, err error) {
	if err != nil {
		log.Errorf(msg, err)
	}
}

func PanicCond(msg string, err error) {
	if err != nil {
		log.Panicf(msg, err)
	}
}
