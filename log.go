package zhanio

import "log"

func sniffError(err error) {
	if err != nil && err != errServerShutdown {
		log.Println(err)
	}
}
