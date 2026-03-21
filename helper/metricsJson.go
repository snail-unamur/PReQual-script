package helper

import (
	"PReQual/model"
	"strconv"
)

func ConvertMeasuresToMap(measures model.SonarMeasures) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range measures.Component.Measures {
		if f, err := strconv.ParseFloat(m.Value, 64); err == nil {
			result[m.Metric] = f
		} else {
			result[m.Metric] = m.Value
		}
	}

	return result
}
