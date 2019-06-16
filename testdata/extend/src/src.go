package src

type Data struct {
	i  int
	is []int
}

func (d *Data) method1() {
	d.i = 0
}

func (d *Data) method2(i int) {
	d.is[i] = 123
}
