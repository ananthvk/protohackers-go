package internal

type price struct {
	timestamp int32
	value     int32
}

type State struct {
	prices []price
}

func (s *State) Insert(timestamp int32, value int32) {
	s.prices = append(s.prices, price{timestamp: timestamp, value: value})
}

// Computes the average price within the closed interval [startTs, endTs], and returns the
// result after truncating it to an integer
func (s *State) QueryAverage(startTs int32, endTs int32) int32 {
	var sum, count int64
	for _, price := range s.prices {
		if price.timestamp >= startTs && price.timestamp <= endTs {
			sum += int64(price.value)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return int32(sum / count)
}

func (s *State) Clear() {
	s.prices = nil
}
