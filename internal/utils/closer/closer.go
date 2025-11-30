package closer

import (
	"io"
	"log"
)

func Close(c io.Closer, l ...bool) {
	if err := c.Close(); err != nil && len(l) > 0 && l[0] == true {
		log.Printf("Error closing %T: %v", c, err)
	}
}
