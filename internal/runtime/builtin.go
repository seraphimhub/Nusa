package runtime

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Builtin adalah fungsi yang disediakan runtime.
type Builtin struct {
	Name string
	Fn   func(args ...Object) Object
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "fungsi bawaan " + b.Name }

func newBuiltin(name string, fn func(args ...Object) Object) *Builtin {
	return &Builtin{Name: name, Fn: fn}
}

func newBuiltins() map[string]*Builtin {
	list := []*Builtin{
		newBuiltin("panjang", builtinPanjang),
		newBuiltin("tipe", builtinTipe),
		newBuiltin("ke_teks", builtinKeTeks),
		newBuiltin("ke_angka", builtinKeAngka),
		newBuiltin("ke_desimal", builtinKeDesimal),
		newBuiltin("huruf_besar", builtinHurufBesar),
		newBuiltin("huruf_kecil", builtinHurufKecil),
		newBuiltin("potong", builtinPotong),
		newBuiltin("cari", builtinCari),
		newBuiltin("pisah", builtinPisah),
		newBuiltin("gabung", builtinGabung),
		newBuiltin("tambahkan", builtinTambahkan),
		newBuiltin("hapus", builtinHapus),
		newBuiltin("urutkan", builtinUrutkan),
		newBuiltin("balik", builtinBalik),
		newBuiltin("jumlah", builtinJumlah),
		newBuiltin("akar", builtinAkar),
		newBuiltin("pangkat", builtinPangkat),
		newBuiltin("bulat", builtinBulat),
		newBuiltin("lantai", builtinLantai),
		newBuiltin("langit", builtinLangit),
		newBuiltin("mutlak", builtinMutlak),
		newBuiltin("rentang", builtinRentang),
	}

	builtins := make(map[string]*Builtin, len(list))
	for _, b := range list {
		builtins[b.Name] = b
	}
	return builtins
}

func builtinPanjang(args ...Object) Object {
	if err := needArgs("panjang", args, 1); err != nil {
		return err
	}

	switch obj := args[0].(type) {
	case *String:
		return &Integer{Value: int64(runeLen(obj.Value))}
	case *List:
		return &Integer{Value: int64(len(obj.Elements))}
	default:
		return builtinError("panjang() hanya menerima teks atau daftar, mendapat %s", typeName(args[0]))
	}
}

func builtinTipe(args ...Object) Object {
	if err := needArgs("tipe", args, 1); err != nil {
		return err
	}
	return &String{Value: typeName(args[0])}
}

func builtinKeTeks(args ...Object) Object {
	if err := needArgs("ke_teks", args, 1); err != nil {
		return err
	}
	if s, ok := args[0].(*String); ok {
		return s
	}
	return &String{Value: args[0].Inspect()}
}

func builtinKeAngka(args ...Object) Object {
	if err := needArgs("ke_angka", args, 1); err != nil {
		return err
	}

	switch obj := args[0].(type) {
	case *Integer:
		return obj
	case *Float:
		return &Integer{Value: int64(obj.Value)}
	case *String:
		value, err := strconv.ParseInt(strings.TrimSpace(obj.Value), 10, 64)
		if err != nil {
			return builtinError("ke_angka() tidak bisa mengubah %q menjadi angka", obj.Value)
		}
		return &Integer{Value: value}
	default:
		return builtinError("ke_angka() tidak bisa mengubah %s menjadi angka", typeName(args[0]))
	}
}

func builtinKeDesimal(args ...Object) Object {
	if err := needArgs("ke_desimal", args, 1); err != nil {
		return err
	}

	switch obj := args[0].(type) {
	case *Integer:
		return &Float{Value: float64(obj.Value)}
	case *Float:
		return obj
	case *String:
		value, err := strconv.ParseFloat(strings.TrimSpace(obj.Value), 64)
		if err != nil {
			return builtinError("ke_desimal() tidak bisa mengubah %q menjadi desimal", obj.Value)
		}
		return &Float{Value: value}
	default:
		return builtinError("ke_desimal() tidak bisa mengubah %s menjadi desimal", typeName(args[0]))
	}
}

func builtinHurufBesar(args ...Object) Object {
	if err := needArgs("huruf_besar", args, 1); err != nil {
		return err
	}
	s, ok := args[0].(*String)
	if !ok {
		return builtinError("huruf_besar() membutuhkan teks, mendapat %s", typeName(args[0]))
	}
	return &String{Value: strings.ToUpper(s.Value)}
}

func builtinHurufKecil(args ...Object) Object {
	if err := needArgs("huruf_kecil", args, 1); err != nil {
		return err
	}
	s, ok := args[0].(*String)
	if !ok {
		return builtinError("huruf_kecil() membutuhkan teks, mendapat %s", typeName(args[0]))
	}
	return &String{Value: strings.ToLower(s.Value)}
}

// potong(teks, mulai[, akhir]) memotong berdasarkan karakter (rune).
func builtinPotong(args ...Object) Object {
	if len(args) != 2 && len(args) != 3 {
		return builtinError("potong() membutuhkan 2 atau 3 argumen, diberikan %d", len(args))
	}

	s, ok := args[0].(*String)
	if !ok {
		return builtinError("potong() membutuhkan teks pada argumen pertama, mendapat %s", typeName(args[0]))
	}

	runes := []rune(s.Value)
	start, err := intArg("potong", args[1])
	if err != nil {
		return err
	}

	end := int64(len(runes))
	if len(args) == 3 {
		end, err = intArg("potong", args[2])
		if err != nil {
			return err
		}
	}

	from := clampIndex(int(start), len(runes))
	to := clampIndex(int(end), len(runes))
	if to < from {
		to = from
	}

	return &String{Value: string(runes[from:to])}
}

func builtinCari(args ...Object) Object {
	if err := needArgs("cari", args, 2); err != nil {
		return err
	}

	text, ok := args[0].(*String)
	if !ok {
		return builtinError("cari() membutuhkan teks pada argumen pertama, mendapat %s", typeName(args[0]))
	}
	sub, ok := args[1].(*String)
	if !ok {
		return builtinError("cari() membutuhkan teks pada argumen kedua, mendapat %s", typeName(args[1]))
	}

	byteIndex := strings.Index(text.Value, sub.Value)
	if byteIndex < 0 {
		return &Integer{Value: -1}
	}
	return &Integer{Value: int64(runeLen(text.Value[:byteIndex]))}
}

func builtinPisah(args ...Object) Object {
	if err := needArgs("pisah", args, 2); err != nil {
		return err
	}

	text, ok := args[0].(*String)
	if !ok {
		return builtinError("pisah() membutuhkan teks pada argumen pertama, mendapat %s", typeName(args[0]))
	}
	separator, ok := args[1].(*String)
	if !ok {
		return builtinError("pisah() membutuhkan teks pada argumen kedua, mendapat %s", typeName(args[1]))
	}

	parts := strings.Split(text.Value, separator.Value)
	elements := make([]Object, len(parts))
	for i, part := range parts {
		elements[i] = &String{Value: part}
	}
	return &List{Elements: elements}
}

func builtinGabung(args ...Object) Object {
	if len(args) != 1 && len(args) != 2 {
		return builtinError("gabung() membutuhkan 1 atau 2 argumen, diberikan %d", len(args))
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("gabung() membutuhkan daftar pada argumen pertama, mendapat %s", typeName(args[0]))
	}

	separator := ""
	if len(args) == 2 {
		s, ok := args[1].(*String)
		if !ok {
			return builtinError("gabung() membutuhkan teks sebagai pemisah, mendapat %s", typeName(args[1]))
		}
		separator = s.Value
	}

	parts := make([]string, len(list.Elements))
	for i, element := range list.Elements {
		if text, ok := element.(*String); ok {
			parts[i] = text.Value
		} else {
			parts[i] = element.Inspect()
		}
	}
	return &String{Value: strings.Join(parts, separator)}
}

func builtinTambahkan(args ...Object) Object {
	if err := needArgs("tambahkan", args, 2); err != nil {
		return err
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("tambahkan() membutuhkan daftar pada argumen pertama, mendapat %s", typeName(args[0]))
	}

	list.Elements = append(list.Elements, args[1])
	return list
}

func builtinHapus(args ...Object) Object {
	if err := needArgs("hapus", args, 2); err != nil {
		return err
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("hapus() membutuhkan daftar pada argumen pertama, mendapat %s", typeName(args[0]))
	}

	index, err := resolveIndex(args[1], len(list.Elements))
	if err != nil {
		return builtinError("hapus(): %s", err.Error())
	}

	list.Elements = append(list.Elements[:index], list.Elements[index+1:]...)
	return list
}

func builtinUrutkan(args ...Object) Object {
	if err := needArgs("urutkan", args, 1); err != nil {
		return err
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("urutkan() membutuhkan daftar, mendapat %s", typeName(args[0]))
	}
	if len(list.Elements) < 2 {
		return list
	}

	switch {
	case isNumeric(list.Elements[0]):
		for _, element := range list.Elements {
			if !isNumeric(element) {
				return builtinError("urutkan() tidak bisa mencampur angka dengan %s", typeName(element))
			}
		}
		sort.SliceStable(list.Elements, func(i, j int) bool {
			left, _ := toFloat64(list.Elements[i])
			right, _ := toFloat64(list.Elements[j])
			return left < right
		})
	case list.Elements[0].Type() == STRING_OBJ:
		for _, element := range list.Elements {
			if element.Type() != STRING_OBJ {
				return builtinError("urutkan() tidak bisa mencampur teks dengan %s", typeName(element))
			}
		}
		sort.SliceStable(list.Elements, func(i, j int) bool {
			return list.Elements[i].(*String).Value < list.Elements[j].(*String).Value
		})
	default:
		return builtinError("urutkan() hanya mendukung daftar angka atau daftar teks, mendapat %s", typeName(list.Elements[0]))
	}

	return list
}

func builtinBalik(args ...Object) Object {
	if err := needArgs("balik", args, 1); err != nil {
		return err
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("balik() membutuhkan daftar, mendapat %s", typeName(args[0]))
	}

	elements := make([]Object, len(list.Elements))
	for i, element := range list.Elements {
		elements[len(list.Elements)-1-i] = element
	}
	return &List{Elements: elements}
}

func builtinJumlah(args ...Object) Object {
	if err := needArgs("jumlah", args, 1); err != nil {
		return err
	}

	list, ok := args[0].(*List)
	if !ok {
		return builtinError("jumlah() membutuhkan daftar angka, mendapat %s", typeName(args[0]))
	}

	var intTotal int64
	var floatTotal float64
	allInts := true

	for _, element := range list.Elements {
		switch value := element.(type) {
		case *Integer:
			intTotal += value.Value
			floatTotal += float64(value.Value)
		case *Float:
			allInts = false
			floatTotal += value.Value
		default:
			return builtinError("jumlah() hanya menerima angka, mendapat %s", typeName(element))
		}
	}

	if allInts {
		return &Integer{Value: intTotal}
	}
	return &Float{Value: floatTotal}
}

func builtinAkar(args ...Object) Object {
	if err := needArgs("akar", args, 1); err != nil {
		return err
	}

	value, ok := toFloat64(args[0])
	if !ok {
		return builtinError("akar() membutuhkan angka, mendapat %s", typeName(args[0]))
	}
	if value < 0 {
		return builtinError("akar() tidak menerima angka negatif")
	}
	return &Float{Value: math.Sqrt(value)}
}

func builtinPangkat(args ...Object) Object {
	if err := needArgs("pangkat", args, 2); err != nil {
		return err
	}

	base, ok := toFloat64(args[0])
	if !ok {
		return builtinError("pangkat() membutuhkan angka pada argumen pertama, mendapat %s", typeName(args[0]))
	}
	exponent, ok := toFloat64(args[1])
	if !ok {
		return builtinError("pangkat() membutuhkan angka pada argumen kedua, mendapat %s", typeName(args[1]))
	}

	result := math.Pow(base, exponent)

	// Pertahankan hasil bulat untuk pangkat bilangan bulat non-negatif.
	if _, baseInt := args[0].(*Integer); baseInt {
		if _, expInt := args[1].(*Integer); expInt && exponent >= 0 &&
			result == math.Trunc(result) && math.Abs(result) <= math.MaxInt64 {
			return &Integer{Value: int64(result)}
		}
	}

	return &Float{Value: result}
}

func builtinBulat(args ...Object) Object {
	return roundBuiltin("bulat", math.Round, args)
}

func builtinLantai(args ...Object) Object {
	return roundBuiltin("lantai", math.Floor, args)
}

func builtinLangit(args ...Object) Object {
	return roundBuiltin("langit", math.Ceil, args)
}

func roundBuiltin(name string, round func(float64) float64, args []Object) Object {
	if err := needArgs(name, args, 1); err != nil {
		return err
	}
	value, ok := toFloat64(args[0])
	if !ok {
		return builtinError("%s() membutuhkan angka, mendapat %s", name, typeName(args[0]))
	}
	return &Integer{Value: int64(round(value))}
}

func builtinMutlak(args ...Object) Object {
	if err := needArgs("mutlak", args, 1); err != nil {
		return err
	}

	switch value := args[0].(type) {
	case *Integer:
		if value.Value < 0 {
			return &Integer{Value: -value.Value}
		}
		return value
	case *Float:
		return &Float{Value: math.Abs(value.Value)}
	default:
		return builtinError("mutlak() membutuhkan angka, mendapat %s", typeName(args[0]))
	}
}

// rentang(n), rentang(mulai, selesai[, langkah]) menghasilkan daftar angka.
func builtinRentang(args ...Object) Object {
	if len(args) < 1 || len(args) > 3 {
		return builtinError("rentang() membutuhkan 1 sampai 3 argumen, diberikan %d", len(args))
	}

	start := int64(0)
	end, err := intArg("rentang", args[0])
	if err != nil {
		return err
	}
	step := int64(1)

	if len(args) >= 2 {
		start = end
		end, err = intArg("rentang", args[1])
		if err != nil {
			return err
		}
	}
	if len(args) == 3 {
		step, err = intArg("rentang", args[2])
		if err != nil {
			return err
		}
	}
	if step == 0 {
		return builtinError("rentang(): langkah tidak boleh nol")
	}

	elements := []Object{}
	for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
		elements = append(elements, &Integer{Value: i})
	}
	return &List{Elements: elements}
}

// --- Bantuan ---

func needArgs(name string, args []Object, count int) *Error {
	if len(args) != count {
		return builtinError("%s() membutuhkan %d argumen, diberikan %d", name, count, len(args))
	}
	return nil
}

func intArg(name string, arg Object) (int64, *Error) {
	value, ok := arg.(*Integer)
	if !ok {
		return 0, builtinError("%s() membutuhkan angka bulat, mendapat %s", name, typeName(arg))
	}
	return value.Value, nil
}

func builtinError(format string, args ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}

func clampIndex(index, length int) int {
	if index < 0 {
		index = length + index
	}
	if index < 0 {
		return 0
	}
	if index > length {
		return length
	}
	return index
}
