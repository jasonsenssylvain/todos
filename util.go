package main

import "reflect"

func Map(f interface{}, vs interface{}) interface{} {

	vf := reflect.ValueOf(f)
	vx := reflect.ValueOf(vs)

	l := vx.Len()

	tys := reflect.SliceOf(vf.Type().Out(0))
	vys := reflect.MakeSlice(tys, l, l)

	for i := 0; i < l; i++ {

		y := vf.Call([]reflect.Value{vx.Index(i)})[0]
		vys.Index(i).Set(y)
	}

	return vys.Interface()
}

func Filter(f interface{}, vs interface{}) interface{} {

	vf := reflect.ValueOf(f)
	vx := reflect.ValueOf(vs)

	l := vx.Len()

	tys := reflect.SliceOf(vf.Type().In(0))

	vss := []reflect.Value{}

	for i := 0; i < l; i++ {

		v := vx.Index(i)
		if vf.Call([]reflect.Value{v})[0].Bool() {

			vss = append(vss, v)
		}
	}

	vys := reflect.MakeSlice(tys, len(vss), len(vss))

	for i, v := range vss {
		vys.Index(i).Set(v)
	}

	return vys.Interface()
}
