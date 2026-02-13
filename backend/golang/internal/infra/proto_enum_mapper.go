package infra

import commonv1 "github.com/itsuabush1003/cursed-frame/backend/golang/internal/gen/common/v1"

func ResultStateMapper(rate float32) int32 {
	switch {
	case rate >= 1.001:	// 計算誤差が出たとき用に少しバッファを設けてる
		// 正解率が１を超えるのはチートか計算ミス
		return int32(commonv1.Result_UNSPECIFIED)
	case rate > 0.999:	// 計算誤差が出たとき用に少しバッファを設けてる
		return int32(commonv1.Result_PERFECT)
	case rate >= 0.9:
		return int32(commonv1.Result_EXCELLENT)
	case rate >= 0.75:
		return int32(commonv1.Result_GREAT)
	case rate >= 0.5:
		return int32(commonv1.Result_GOODJOB)
	case rate >= 0.3:
		return int32(commonv1.Result_CLEAR)
	case rate >= 0.0:
		return int32(commonv1.Result_FAILED)
	default:
		// 正解率が０を下回るのはチートか計算ミス
		return int32(commonv1.Result_UNSPECIFIED)
	}
}