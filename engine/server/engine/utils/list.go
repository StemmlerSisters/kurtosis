package utils

func MapList[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))
	for _, e := range data {
		res = append(res, f(e))
	}
	return res
}

func MapListStopOnError[T, U any](data []T, f func(T) (U, error)) ([]U, error) {
	res := make([]U, 0, len(data))
	for _, e := range data {
		y, err := f(e)
		if err != nil {
			return res, err
		}
		res = append(res, y)
	}
	return res, nil
}
