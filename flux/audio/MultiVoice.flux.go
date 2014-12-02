package audio

import (
	"sync"
)

type MultiVoice struct{Params Params; voices map[Voice]struct{}; mu sync.Mutex; Out Audio}