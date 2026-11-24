package drisl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"

	"codeberg.org/reiver/go-erorr"
)

// ToJSON renders 'value' (as returned by [Unmarshal]) as JSON, using the AT-Protocol conventions:
//
//	• a byte string becomes {"$bytes": "<base64, standard alphabet, no padding>"},
//	• a CID becomes {"$link": "<CID string>"}.
//
// Map keys come out in byte-wise order. HTML characters (<, >, &) are not escaped.
//
// Example usage:
//
//	value, err := drisl.Unmarshal(p)
//	if nil != err {
//		return err
//	}
//
//	j, err := drisl.ToJSON(value)
func ToJSON(value any) ([]byte, error) {
	converted, err := toJSONValue(value, 0)
	if nil != err {
		return nil, err
	}

	var buffer bytes.Buffer
	{
		var encoder *json.Encoder = json.NewEncoder(&buffer)
		encoder.SetEscapeHTML(false)

		err := encoder.Encode(converted)
		if nil != err {
			return nil, erorr.Errorf("drisl: problem rendering JSON: %w", err)
		}
	}

	// json.Encoder adds a trailing newline.
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

func toJSONValue(value any, depth int) (any, error) {
	if MaxDepth < depth {
		return nil, ErrTooDeep
	}

	switch casted := value.(type) {
	case nil, bool, string,
	     int, int8, int16, int32, int64,
	     uint, uint8, uint16, uint32, uint64:
		return casted, nil

	case float32:
		if _, err := marshalFloat64(float64(casted)); nil != err {
			return nil, err
		}
		return casted, nil
	case float64:
		if _, err := marshalFloat64(casted); nil != err {
			return nil, err
		}
		return casted, nil

	case *big.Int:
		if nil == casted {
			return nil, erorr.Errorf("drisl: cannot render a nil *big.Int: %w", ErrUnsupportedType)
		}
		return json.Number(casted.String()), nil

	case []byte:
		return map[string]any{
			"$bytes": base64.RawStdEncoding.EncodeToString(casted),
		}, nil

	case CID:
		if err := casted.Validate(); nil != err {
			return nil, invalidCID(err)
		}
		return map[string]any{
			"$link": casted.String(),
		}, nil

	case []any:
		var result []any = make([]any, 0, len(casted))
		for index, item := range casted {
			converted, err := toJSONValue(item, depth+1)
			if nil != err {
				return nil, erorr.Errorf("drisl: array item [%d]: %w", index, err)
			}
			result = append(result, converted)
		}
		return result, nil

	case map[string]any:
		var result map[string]any = make(map[string]any, len(casted))
		for key, item := range casted {
			converted, err := toJSONValue(item, depth+1)
			if nil != err {
				return nil, erorr.Errorf("drisl: map value for key %q: %w", key, err)
			}
			result[key] = converted
		}
		return result, nil

	default:
		return nil, erorr.Errorf("drisl: cannot render value of type %T as JSON: %w", value, ErrUnsupportedType)
	}
}
