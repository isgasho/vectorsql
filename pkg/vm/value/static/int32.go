package static

import (
	"reflect"
	"unsafe"

	"github.com/pilosa/pilosa/roaring"
)

func NewInt32s(vs []int32, np, dp *roaring.Bitmap) *Int32s {
	return &Int32s{
		Vs: vs,
		Np: np,
		Dp: dp,
	}
}

func (a *Int32s) Size() int {
	return len(a.Vs) * 4
}

func (a *Int32s) Show() ([]byte, error) {
	v, err := show(a.Np)
	if err != nil {
		return nil, err
	}
	hp := *(*reflect.SliceHeader)(unsafe.Pointer(&a.Vs))
	hp.Len *= 4
	hp.Cap *= 4
	return append(v, *(*[]byte)(unsafe.Pointer(&hp))...), nil
}

func (a *Int32s) Read(cnt int, data []byte) error {
	data, np, err := read(data)
	if err != nil {
		return err
	}
	a.Np = np
	hp := *(*reflect.SliceHeader)(unsafe.Pointer(&data))
	hp.Len = cnt
	hp.Cap = cnt
	a.Vs = *(*[]int32)(unsafe.Pointer(&hp))
	return nil
}

func (a *Int32s) MarkNull(row int) error {
	a.Np.DirectAdd(uint64(row))
	return nil
}

func (a *Int32s) Append(v interface{}) error {
	a.Vs = append(a.Vs, v.([]int32)...)
	return nil
}

func (a *Int32s) Merge(np, dp *roaring.Bitmap) error {
	a.Dp = dp
	if a.Np == nil {
		a.Np = np
		return nil
	}
	a.Np = a.Np.Union(np)
	return nil
}

func (a *Int32s) Update(rows []int, v interface{}) error {
	vs := v.([]int32)
	for _, i := range rows {
		a.Vs[i] = vs[i]
	}
	return nil
}

func (a *Int32s) Filter(is []uint64) interface{} {
	if len(is) == 0 {
		return &Int32s{}
	}
	return &Int32s{
		Is: is,
		Vs: a.Vs,
		Np: a.Np,
		Dp: a.Dp,
	}
}

func (a *Int32s) MergeFilter(v interface{}) interface{} {
	b := v.(*Bools)
	r := &Int32s{
		Vs: a.Vs,
	}
	switch {
	case a.Np != nil && b.Np == nil:
		r.Np = a.Np
	case a.Np == nil && b.Np != nil:
		r.Np = b.Np
	case a.Np != nil && b.Np != nil:
		r.Np = a.Np.Union(b.Np)
	}
	switch {
	case a.Dp != nil && b.Dp == nil:
		r.Dp = a.Dp
	case a.Dp == nil && b.Dp != nil:
		r.Dp = b.Dp
	case a.Dp != nil && b.Dp != nil:
		r.Dp = a.Np.Union(b.Dp)
	}
	switch {
	case len(a.Is) > 0 && len(b.Is) > 0:
		mp := make(map[uint64]struct{})
		{
			for _, o := range a.Is {
				mp[o] = struct{}{}
			}
		}
		r.Is = make([]uint64, 0, len(b.Is))
		for _, o := range b.Is {
			if _, ok := mp[o]; ok && b.Vs[o] {
				r.Is = append(r.Is, o)
			}
		}
	case len(a.Is) > 0 && len(b.Is) == 0:
		r.Is = make([]uint64, 0, len(a.Is))
		for _, o := range a.Is {
			if b.Vs[o] {
				r.Is = append(r.Is, o)
			}
		}
	case len(a.Is) == 0 && len(b.Is) > 0:
		r.Is = make([]uint64, 0, len(b.Is))
		for _, o := range b.Is {
			if b.Vs[o] {
				r.Is = append(r.Is, o)
			}
		}
	case len(a.Is) == 0 && len(b.Is) == 0:
		r.Is = make([]uint64, 0, len(a.Vs))
		for i := range a.Vs {
			if b.Vs[i] {
				r.Is = append(r.Is, uint64(i))
			}
		}
	}
	return r
}
