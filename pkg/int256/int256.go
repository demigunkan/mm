// Package int256 provides arithmatic functions fixed size 256-bit math
// A fork of https://github.com/holiman/uint256
// All credits to the author of https://github.com/holiman/uint256

package int256

import (
	"errors"
	"io"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

// Int is represented as an array of 4+2 uint64, in little-endian order,
// so that Int[3] is the most significant, and Int[0] is the least significant
// The 5th element stores the sign (1 if negative) and the last element stores the decimal place
type Int [6]uint64

//////////////// ERRORS

var (
	ErrBig256Range      = errors.New("hex number > 256 bits")
	ErrEmptyString      = errors.New("empty string")
	ErrNotDecimalString = errors.New("not decimal string")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrOverflow         = errors.New("overflow")
	ErrNil              = errors.New("nil")
)

//////////////// CONSTANTS

const (
	defaultDecimal uint64 = 6
)

var (
	O, _       = (Int{}).FromInt(0, int(defaultDecimal))
	I, _       = (Int{}).FromInt(1, int(defaultDecimal))
	MaxUint256 = Int{_max_u64, _max_u64, _max_u64, _max_u64, _false, defaultDecimal}
)

//////////////// NEW INSTANCE

func (n Int) SetNil() Int {
	n = Int{}
	return n
}

func (n Int) SetMax() Int {
	n = MaxUint256
	return n
}

func (n Int) FromInt(val int64, dec int) (Int, error) {
	multiplier := newInt(uint64(math.Pow10(int(dec))))

	neg := _false
	if val < 0 {
		neg = _true
		val = -val
	}

	n.setUint64(uint64(val))
	if _, overflow := n.mul(&n, multiplier); overflow {
		return n, ErrOverflow
	}

	n[_neg] = neg
	n[_dec] = uint64(dec)

	return n, nil
}

func (n Int) FromBigInt(val *big.Int, dec int) (Int, error) {
	if val == nil {
		return n, ErrNil
	}

	neg := _false
	if val.Cmp(_zero_big) == -1 {
		val = new(big.Int).Abs(val)
		neg = _true
	}

	n.setFromBig(val)
	n[_dec] = uint64(dec)
	n[_neg] = neg

	return n, nil
}

func (n Int) FromString(val string, dec int) (Int, error) {
	if len(val) == 0 {
		return n, ErrEmptyString
	}
	neg := _false
	if val[0] == '-' {
		val = val[1:]
		neg = _true
	}

	if err := n.setFromDecimal(val); err != nil {
		return n, err
	}
	n[_dec] = uint64(dec)
	n[_neg] = neg

	return n, nil
}

func (n Int) FromFloat(val float64, dec int) (Int, error) {
	str := strconv.FormatFloat(val, 'f', -1, 64)
	return n.FromText(str, dec)
}

func (n Int) FromText(val string, dec int) (Int, error) {
	if len(val) == 0 {
		return n, ErrEmptyString
	}

	if val == _zero {
		return n.FromString(val, dec)
	}

	i := strings.Index(val, _dot)

	var str string
	var decimal string
	if i < 0 {
		str = val
	} else {
		decimal = val[i+1:]
		str = val[:i] + decimal
	}

	neg := _false
	if str[0] == '-' {
		str = str[1:]
		neg = _true
	}

	for str[0] == '0' {
		str = str[1:]
	}

	if err := n.setFromDecimal(str); err != nil {
		return n, err
	}

	if d := dec - len(decimal); d < 0 {
		delta := uint64(math.Pow10(int(-d)))
		divisor := newInt(uint64(math.Pow10(int(-d))))
		(&n).add(&n, newInt(delta/2))
		(&n).div(&n, divisor)
	} else {
		n.mul(&n, newInt(uint64(math.Pow10(int(d)))))
	}

	n[_dec] = uint64(dec)
	n[_neg] = neg

	return n, nil
}

//////////////// NUM INTERFACE

// WARN: we assume z is nil if it has zero decimals
// This is a temporary solution to fit into the Num interface
func (z Int) IsNil() bool {
	return z[_dec] == 0
}

func (z Int) Float(dec int) float64 {
	str := z.Text(dec)
	f, _ := strconv.ParseFloat(str, 64)

	return f
}

func (z Int) BigInt() *big.Int {
	return (&z).toBig()
}

func (z Int) String() string {
	if z[_neg] == _true {
		return _neg_sign + z.dec()
	}
	return z.dec()
}

func (z Int) Text(dec int) string {
	str := z.dec()
	if str == _zero {
		return str
	}
	if str == _max_u64_str {
		return _max_u64_text[:len(_max_u64_text)-dec]
	}

	var text string

	ln := len(str)
	if ln <= dec {
		text = _zero + _dot + strings.Repeat(_zero, dec-ln) + str
	} else {
		text = str[:ln-dec] + _dot + str[ln-dec:]
	}

	ln = len(text) - 1
	for i := range text[:] {
		if text[ln-i] == '0' {
			text = text[:ln-i]
			continue
		}
		if text[ln-i] == '.' {
			text = text[:ln-i]
			break
		}
		break
	}

	if z[_neg] == _true {
		return _neg_sign + text
	}

	return text
}

func (x Int) Add(y Int) (z Int) {
	if x[_dec] != y[_dec] {
		normalize(&x, &y)
	}

	neg := x[_neg]

	if x[_neg] == y[_neg] {
		// x + y == x + y
		// (-x) + (-y) == -(x + y)
		if _, overflow := (&z).add(&x, &y); overflow {
			panic(ErrOverflow)
		}
	} else {
		// x + (-y) == x - y == -(y - x)
		// (-x) + y == y - x == -(x - y)
		if x.cmp(&y) >= 0 {
			if _, overflow := z.sub(&x, &y); overflow {
				panic(ErrOverflow)
			}
		} else {
			neg = 1 - neg
			if _, overflow := (&z).sub(&y, &x); overflow {
				panic(ErrOverflow)
			}
		}
	}
	z[_dec] = x[_dec] // TO DO: decimal adjustment
	z[_neg] = neg     // 0 has no sign

	return
}

func (x Int) Sub(y Int) (z Int) {
	if x[_dec] != y[_dec] {
		normalize(&x, &y)
	}

	neg := x[_neg]
	if x[_neg] != y[_neg] {
		// x - (-y) == x + y
		// (-x) - y == -(x + y)
		if _, overflow := (&z).add(&x, &y); overflow {
			panic(ErrOverflow)
		}
	} else {
		// x - y == x - y == -(y - x)
		// (-x) - (-y) == y - x == -(x - y)
		if x.cmp(&y) >= 0 {
			if _, overflow := z.sub(&x, &y); overflow {
				panic(ErrOverflow)
			}
		} else {
			neg = 1 - neg
			if _, overflow := (&z).sub(&y, &x); overflow {
				panic(ErrOverflow)
			}
		}
	}
	z[_dec] = x[_dec] // TO DO: decimal adjustment
	z[_neg] = neg     // 0 has no sign
	return z
}

func (x Int) Div(y Int) (z Int) {
	if x[_dec] != y[_dec] {
		normalize(&x, &y)
	}

	dec := uint64(math.Pow10(int(x[_dec])))
	multiplier := newInt(dec)
	multiplier[_dec] = x[_dec]
	if _, overflow := (&z).mul(&x, multiplier); overflow {
		panic(ErrOverflow)
	}

	(&z).add(&z, new(Int).div(&y, newInt(2)))
	(&z).div(&z, &y)

	z[_dec] = x[_dec] // TO DO: decimal adjustment
	z[_neg] = (x[_neg] + y[_neg]) % 2

	return
}

func (x Int) RDiv(y Int) (z Int) {
	(&z).div(&x, &y)
	z[_dec] = x[_dec]                 // TO DO: decimal adjustment
	z[_neg] = (x[_neg] + y[_neg]) % 2 // 0 has no sign
	return z
}

func (x Int) Mul(y Int) (z Int) {
	if x[_dec] != y[_dec] {
		normalize(&x, &y)
	}

	if _, overflow := (&z).mul(&x, &y); overflow {
		panic(ErrOverflow)
	}

	dec := uint64(math.Pow10(int(x[_dec])))
	multiplier := newInt(dec)

	(&z).add(&z, newInt(dec/2))
	(&z).div(&z, multiplier)

	z[_dec] = x[_dec] // TO DO: decimal adjustment
	z[_neg] = (x[_neg] + y[_neg]) % 2

	return
}

func (x Int) RMul(y Int) (z Int) {
	if _, overflow := (&z).mul(&x, &y); overflow {
		panic(ErrOverflow)
	}

	z[_dec] = x[_dec]                 // TO DO: decimal adjustment
	z[_neg] = (x[_neg] + y[_neg]) % 2 // 0 has no sign
	return z
}

func (x Int) Mod(y Int) (z Int) {
	(&z).mod(&x, &y)
	z[_dec] = x[_dec]                 // TO DO: decimal adjustment
	z[_neg] = (x[_neg] + y[_neg]) % 2 // 0 has no sign
	return z
}

func (x Int) Abs() Int {
	x[_neg] = _false
	return x
}

func (x Int) Neg() Int {
	return x.Negate()
}

func (x Int) NegIf(flag bool) Int {
	if flag {
		x[_neg] = _true
	}
	return x
}

func (x Int) Negate() Int {
	x[_neg] = 1 - x[_neg]
	return x
}

func (x Int) Round(step Int, up bool) (z Int) {
	// TO DO: normalize decimals
	m := new(Int)
	z.divmod(&x, &step, m)

	if m.isZero() {
		return x
	}

	z.mul(&z, &step)
	if up {
		z.add(&z, &step)
	}
	z[_neg] = x[_neg]
	z[_dec] = x[_dec]

	return
}

func (x Int) Sign() int {
	switch {
	case x.isZero():
		return 0
	case x[_neg] == _true:
		return -1
	default:
		return 1
	}
}

func (x Int) Cmp(y Int) (r int) {
	switch {
	case x[_neg] == y[_neg]:
		if x[_dec] != y[_dec] {
			normalize(&x, &y)
		}
		r = (&x).cmp(&y)
		if x[_neg] == _true {
			r = -r
		}
	case x[_neg] == _true:
		r = -1
	default:
		r = 1
	}

	return
}

func (x Int) Lt(y Int) (r bool) {
	switch {
	case x[_neg] == y[_neg]:
		if x[_dec] != y[_dec] {
			normalize(&x, &y)
		}
		r = (&x).lt(&y)
		if x[_neg] == _true {
			r = !r
		}
	case x[_neg] == _true:
		r = true
	default:
		r = false
	}

	return
}

func (x Int) Equal(y Int) (r bool) {
	if x[_dec] != y[_dec] {
		normalize(&x, &y)
	}

	return x.eq(&y)
}

//////////////// INTERNAL

// Internal constants
const (
	_neg          = 4
	_dec          = 5
	_true         = uint64(1)
	_false        = uint64(0)
	_zero         = "0"
	_dot          = "."
	_neg_sign     = "-"
	_max_u64      = math.MaxUint64
	_max_u64_str  = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
	_max_u64_text = "115792089237316200000000000000000000000000000000000000000000000000000000000000"
	_max_words    = 256 / bits.UintSize
)

var (
	_zero_big = big.NewInt(0)
)

func newInt(val uint64) *Int {
	z := &Int{}
	z.setUint64(val)
	return z
}

func normalize(x, y *Int) {
	if x[_dec] > y[_dec] {
		y.upscale(x[_dec])
	} else {
		x.upscale(y[_dec])
	}
}

func (z *Int) upscale(dec uint64) *Int {
	neg := z[_neg]
	multiplier := newInt(uint64(math.Pow10(int(dec - z[_dec]))))
	z.mul(z, multiplier)
	z[_neg] = neg
	z[_dec] = dec

	return z
}

// Uint256 operation: Add sets z to the sum x+y
func (z *Int) add(x, y *Int) (*Int, bool) {
	var carry uint64
	z[0], carry = bits.Add64(x[0], y[0], 0)
	z[1], carry = bits.Add64(x[1], y[1], carry)
	z[2], carry = bits.Add64(x[2], y[2], carry)
	z[3], carry = bits.Add64(x[3], y[3], carry)
	return z, carry != 0
}

// Uint256 operation: Sub sets z to the difference x-y
func (z *Int) sub(x, y *Int) (*Int, bool) {
	var carry uint64
	z[0], carry = bits.Sub64(x[0], y[0], 0)
	z[1], carry = bits.Sub64(x[1], y[1], carry)
	z[2], carry = bits.Sub64(x[2], y[2], carry)
	z[3], _ = bits.Sub64(x[3], y[3], carry)
	return z, carry != 0
}

// Cmp compares z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *Int) cmp(x *Int) (r int) {
	if z.gt(x) {
		return 1
	}
	if z.lt(x) {
		return -1
	}
	return 0
}

// Gt returns true if z > x
func (z *Int) gt(x *Int) bool {
	return x.lt(z)
}

// Lt returns true if z < x
func (z Int) lt(x *Int) bool {
	// z < x <=> z - x < 0 i.e. when subtraction overflows.
	_, carry := bits.Sub64(z[0], x[0], 0)
	_, carry = bits.Sub64(z[1], x[1], carry)
	_, carry = bits.Sub64(z[2], x[2], carry)
	_, carry = bits.Sub64(z[3], x[3], carry)
	return carry != 0
}

// umul computes full 256 x 256 -> 512 multiplication.
func umul(x, y *Int) [8]uint64 {
	var (
		res                           [8]uint64
		carry, carry4, carry5, carry6 uint64
		res1, res2, res3, res4, res5  uint64
	)

	carry, res[0] = bits.Mul64(x[0], y[0])
	carry, res1 = umulHop(carry, x[1], y[0])
	carry, res2 = umulHop(carry, x[2], y[0])
	carry4, res3 = umulHop(carry, x[3], y[0])

	carry, res[1] = umulHop(res1, x[0], y[1])
	carry, res2 = umulStep(res2, x[1], y[1], carry)
	carry, res3 = umulStep(res3, x[2], y[1], carry)
	carry5, res4 = umulStep(carry4, x[3], y[1], carry)

	carry, res[2] = umulHop(res2, x[0], y[2])
	carry, res3 = umulStep(res3, x[1], y[2], carry)
	carry, res4 = umulStep(res4, x[2], y[2], carry)
	carry6, res5 = umulStep(carry5, x[3], y[2], carry)

	carry, res[3] = umulHop(res3, x[0], y[3])
	carry, res[4] = umulStep(res4, x[1], y[3], carry)
	carry, res[5] = umulStep(res5, x[2], y[3], carry)
	res[7], res[6] = umulStep(carry6, x[3], y[3], carry)

	return res
}

// Mul sets z to the product x*y
func (z *Int) mul(x, y *Int) (*Int, bool) {
	p := umul(x, y)
	copy(z[:], p[:4])
	return z, (p[4] | p[5] | p[6] | p[7]) != 0
}

// Eq returns true if z == x
func (z *Int) eq(x *Int) bool {
	return (z[0] == x[0]) && (z[1] == x[1]) && (z[2] == x[2]) && (z[3] == x[3]) && (z[4] == x[4]) && (z[5] == x[5])
}

// SetOne sets z to 1
func (z *Int) setOne() *Int {
	z[3], z[2], z[1], z[0] = 0, 0, 0, 1
	return z
}

// umulStep computes (hi * 2^64 + lo) = z + (x * y) + carry.
func umulStep(z, x, y, carry uint64) (hi, lo uint64) {
	hi, lo = bits.Mul64(x, y)
	lo, carry = bits.Add64(lo, carry, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	lo, carry = bits.Add64(lo, z, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

// umulHop computes (hi * 2^64 + lo) = z + (x * y)
func umulHop(z, x, y uint64) (hi, lo uint64) {
	hi, lo = bits.Mul64(x, y)
	lo, carry := bits.Add64(lo, z, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

// addTo computes x += y.
// Requires len(x) >= len(y).
func addTo(x, y []uint64) uint64 {
	var carry uint64
	for i := 0; i < len(y); i++ {
		x[i], carry = bits.Add64(x[i], y[i], carry)
	}
	return carry
}

// subMulTo computes x -= y * multiplier.
// Requires len(x) >= len(y).
func subMulTo(x, y []uint64, multiplier uint64) uint64 {

	var borrow uint64
	for i := 0; i < len(y); i++ {
		s, carry1 := bits.Sub64(x[i], borrow, 0)
		ph, pl := bits.Mul64(y[i], multiplier)
		t, carry2 := bits.Sub64(s, pl, 0)
		x[i] = t
		borrow = ph + carry1 + carry2
	}
	return borrow
}

// udivremBy1 divides u by single normalized word d and produces both quotient and remainder.
// The quotient is stored in provided quot.
func udivremBy1(quot, u []uint64, d uint64) (rem uint64) {
	reciprocal := reciprocal2by1(d)
	rem = u[len(u)-1] // Set the top word as remainder.
	for j := len(u) - 2; j >= 0; j-- {
		quot[j], rem = udivrem2by1(rem, u[j], d, reciprocal)
	}
	return rem
}

// udivremKnuth implements the division of u by normalized multiple word d from the Knuth's division algorithm.
// The quotient is stored in provided quot - len(u)-len(d) words.
// Updates u to contain the remainder - len(d) words.
func udivremKnuth(quot, u, d []uint64) {
	dh := d[len(d)-1]
	dl := d[len(d)-2]
	reciprocal := reciprocal2by1(dh)

	for j := len(u) - len(d) - 1; j >= 0; j-- {
		u2 := u[j+len(d)]
		u1 := u[j+len(d)-1]
		u0 := u[j+len(d)-2]

		var qhat, rhat uint64
		if u2 >= dh { // Division overflows.
			qhat = ^uint64(0)
			// TODO: Add "qhat one to big" adjustment (not needed for correctness, but helps avoiding "add back" case).
		} else {
			qhat, rhat = udivrem2by1(u2, u1, dh, reciprocal)
			ph, pl := bits.Mul64(qhat, dl)
			if ph > rhat || (ph == rhat && pl > u0) {
				qhat--
				// TODO: Add "qhat one to big" adjustment (not needed for correctness, but helps avoiding "add back" case).
			}
		}

		// Multiply and subtract.
		borrow := subMulTo(u[j:], d, qhat)
		u[j+len(d)] = u2 - borrow
		if u2 < borrow { // Too much subtracted, add back.
			qhat--
			u[j+len(d)] += addTo(u[j:], d)
		}

		quot[j] = qhat // Store quotient digit.
	}
}

// udivrem divides u by d and produces both quotient and remainder.
// The quotient is stored in provided quot - len(u)-len(d)+1 words.
// It loosely follows the Knuth's division algorithm (sometimes referenced as "schoolbook" division) using 64-bit words.
// See Knuth, Volume 2, section 4.3.1, Algorithm D.
func udivrem(quot, u []uint64, d *Int) (rem Int) {
	var dLen int
	for i := len(d) - 2 - 1; i >= 0; i-- {
		if d[i] != 0 {
			dLen = i + 1
			break
		}
	}

	shift := uint(bits.LeadingZeros64(d[dLen-1]))

	var dnStorage Int
	dn := dnStorage[:dLen]
	for i := dLen - 1; i > 0; i-- {
		dn[i] = (d[i] << shift) | (d[i-1] >> (64 - shift))
	}
	dn[0] = d[0] << shift

	var uLen int
	for i := len(u) - 1; i >= 0; i-- {
		if u[i] != 0 {
			uLen = i + 1
			break
		}
	}

	var unStorage [9]uint64
	un := unStorage[:uLen+1]
	un[uLen] = u[uLen-1] >> (64 - shift)
	for i := uLen - 1; i > 0; i-- {
		un[i] = (u[i] << shift) | (u[i-1] >> (64 - shift))
	}
	un[0] = u[0] << shift

	// TODO: Skip the highest word of numerator if not significant.

	if dLen == 1 {
		r := udivremBy1(quot, un, dn[0])
		rem.setUint64(r >> shift)
		return rem
	}

	udivremKnuth(quot, un, dn)

	for i := 0; i < dLen-1; i++ {
		rem[i] = (un[i] >> shift) | (un[i+1] << (64 - shift))
	}
	rem[dLen-1] = un[dLen-1] >> shift

	return rem
}

// reciprocal2by1 computes <^d, ^0> / d.
func reciprocal2by1(d uint64) uint64 {
	reciprocal, _ := bits.Div64(^d, ^uint64(0), d)
	return reciprocal
}

// udivrem2by1 divides <uh, ul> / d and produces both quotient and remainder.
// It uses the provided d's reciprocal.
// Implementation ported from https://github.com/chfast/intx and is based on
// "Improved division by invariant integers", Algorithm 4.
func udivrem2by1(uh, ul, d, reciprocal uint64) (quot, rem uint64) {
	qh, ql := bits.Mul64(reciprocal, uh)
	ql, carry := bits.Add64(ql, ul, 0)
	qh, _ = bits.Add64(qh, uh, carry)
	qh++

	r := ul - qh*d

	if r > ql {
		qh--
		r += d
	}

	if r >= d {
		qh++
		r -= d
	}

	return qh, r
}

// SetUint64 sets z to the value x
func (z *Int) setUint64(x uint64) *Int {
	z[3], z[2], z[1], z[0] = 0, 0, 0, x
	return z
}

// IsUint64 reports whether z can be represented as a uint64.
func (z *Int) isUint64() bool {
	return (z[1] | z[2] | z[3]) == 0
}

// Clear sets z to 0
func (z *Int) clear() *Int {
	z[5], z[4], z[3], z[2], z[1], z[0] = 0, 0, 0, 0, 0, 0
	return z
}

// Set sets z to x and returns z.
func (z *Int) set(x *Int) *Int {
	*z = *x
	return z
}

// Neg returns -x mod 2**256.
func (z *Int) neg(x *Int) *Int {
	z, _ = z.sub(new(Int), x)
	return z
}

// ToBig returns a big.Int version of z.
func (z *Int) toBig() *big.Int {
	b := new(big.Int)
	if z.isZero() {
		return b
	}
	switch _max_words { // Compile-time check.
	case 4: // 64-bit architectures.
		words := [4]big.Word{big.Word(z[0]), big.Word(z[1]), big.Word(z[2]), big.Word(z[3])}
		b.SetBits(words[:])
	case 8: // 32-bit architectures.
		words := [8]big.Word{
			big.Word(z[0]), big.Word(z[0] >> 32),
			big.Word(z[1]), big.Word(z[1] >> 32),
			big.Word(z[2]), big.Word(z[2] >> 32),
			big.Word(z[3]), big.Word(z[3] >> 32),
		}
		b.SetBits(words[:])
	}

	if z[_neg] == _true {
		b.Neg(b)
	}

	return b
}

func (z *Int) isZero() bool {
	return (z[0] | z[1] | z[2] | z[3]) == 0
}

// SetFromBig converts a big.Int to Int and sets the value to z.
// TODO: Ensure we have sufficient testing, esp for negative bigints.
func (z *Int) setFromBig(b *big.Int) bool {
	z.clear()
	words := b.Bits()
	overflow := len(words) > _max_words

	switch _max_words { // Compile-time check.
	case 4: // 64-bit architectures.
		if len(words) > 0 {
			z[0] = uint64(words[0])
			if len(words) > 1 {
				z[1] = uint64(words[1])
				if len(words) > 2 {
					z[2] = uint64(words[2])
					if len(words) > 3 {
						z[3] = uint64(words[3])
					}
				}
			}
		}
	case 8: // 32-bit architectures.
		numWords := len(words)
		if overflow {
			numWords = _max_words
		}
		for i := 0; i < numWords; i++ {
			if i%2 == 0 {
				z[i/2] = uint64(words[i])
			} else {
				z[i/2] |= uint64(words[i]) << 32
			}
		}
	}

	if b.Sign() == -1 {
		z.neg(z)
	}
	return overflow
}

// Uint64 returns the lower 64-bits of z
func (z *Int) uint64() uint64 {
	return z[0]
}

// SetFromDecimal sets z from the given string, interpreted as a decimal number.
// OBS! This method is _not_ strictly identical to the (*big.Int).SetString(..., 10) method.
// Notable differences:
// - This method does not accept underscore input, e.g. "100_000",
// - This method does not accept negative zero as valid, e.g "-0",
//   - (this method does not accept any negative input as valid))
func (z *Int) setFromDecimal(s string) (err error) {
	// Remove max one leading +
	if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	// Remove any number of leading zeroes
	if len(s) > 0 && s[0] == '0' {
		var i int
		var c rune
		for i, c = range s {
			if c != '0' {
				break
			}
		}
		s = s[i:]
	}
	if len(s) < len(_max_u64_str) {
		return z.fromDecimal(s)
	}
	if len(s) == len(_max_u64_str) {
		if s > _max_u64_str {
			return ErrBig256Range
		}
		return z.fromDecimal(s)
	}
	return ErrBig256Range
}

// multipliers holds the values that are needed for fromDecimal
var multipliers = [5]*Int{
	nil,                             // represents first round, no multiplication needed
	{10000000000000000000, 0, 0, 0}, // 10 ^ 19
	{687399551400673280, 5421010862427522170, 0, 0},                     // 10 ^ 38
	{5332261958806667264, 17004971331911604867, 2938735877055718769, 0}, // 10 ^ 57
	{0, 8607968719199866880, 532749306367912313, 1593091911132452277},   // 10 ^ 76
}

// fromDecimal is a helper function to only ever be called via SetFromDecimal
// this function takes a string and chunks it up, calling ParseUint on it up to 5 times
// these chunks are then multiplied by the proper power of 10, then added together.
func (z *Int) fromDecimal(bs string) error {
	// first clear the input
	z.clear()
	// the maximum value of uint64 is 18446744073709551615, which is 20 characters
	// one less means that a string of 19 9's is always within the uint64 limit
	var (
		num       uint64
		err       error
		remaining = len(bs)
	)
	if remaining == 0 {
		return io.EOF
	}
	// We proceed in steps of 19 characters (nibbles), from least significant to most significant.
	// This means that the first (up to) 19 characters do not need to be multiplied.
	// In the second iteration, our slice of 19 characters needs to be multipleied
	// by a factor of 10^19. Et cetera.
	for i, mult := range multipliers {
		if remaining <= 0 {
			return nil // Done
		} else if remaining > 19 {
			num, err = strconv.ParseUint(bs[remaining-19:remaining], 10, 64)
		} else {
			// Final round
			num, err = strconv.ParseUint(bs, 10, 64)
		}
		if err != nil {
			return err
		}
		// add that number to our running total
		if i == 0 {
			z.setUint64(num)
		} else {
			base := newInt(num)
			base.mul(base, mult)
			z.add(z, base)
		}
		// Chop off another 19 characters
		if remaining > 19 {
			bs = bs[0 : remaining-19]
		}
		remaining -= 19
	}
	return nil
}

func (z *Int) divmod(x, y, m *Int) (*Int, *Int) {
	if x.isZero() || y.isZero() {
		return z.clear(), m.clear()
	}
	var quot Int
	*m = udivrem(quot[:4], x[:4], y)
	*z = quot
	return z, m
}

func (z *Int) div(x, y *Int) *Int {
	if y.isZero() {
		panic(ErrDivisionByZero)
	}

	if y.gt(x) {
		return z.clear()
	}
	if x.eq(y) {
		return z.setOne()
	}
	// Shortcut some cases
	if x.isUint64() {
		return z.setUint64(x.uint64() / y.uint64())
	}

	// At this point, we know
	// x/y ; x > y > 0

	var quot Int
	udivrem(quot[:4], x[:4], y)
	return z.set(&quot)
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, z is set to 0 (OBS: differs from the big.Int)
func (z *Int) mod(x, y *Int) *Int {
	if x.isZero() || y.isZero() {
		return z.clear()
	}
	switch x.cmp(y) {
	case -1:
		// x < y
		copy(z[:], x[:])
		return z
	case 0:
		// x == y
		return z.clear() // They are equal
	}

	// At this point:
	// x != 0
	// y != 0
	// x > y

	// Shortcut trivial case
	if x.isUint64() {
		return z.setUint64(x.uint64() % y.uint64())
	}

	var quot Int
	*z = udivrem(quot[:4], x[:4], y)
	return z
}

// Dec returns the decimal representation of z.
func (z *Int) dec() string {
	if z.isZero() {
		return "0"
	}
	if z.isUint64() {
		return strconv.FormatUint(z.uint64(), 10)
	}
	// The max uint64 value being 18446744073709551615, the largest
	// power-of-ten below that is 10000000000000000000.
	// When we do a DivMod using that number, the remainder that we
	// get back is the lower part of the output.
	//
	// The ascii-output of remainder will never exceed 19 bytes (since it will be
	// below 10000000000000000000).
	//
	// Algorithm example using 100 as divisor
	//
	// 12345 % 100 = 45   (rem)
	// 12345 / 100 = 123  (quo)
	// -> output '45', continue iterate on 123
	var (
		// out is 98 bytes long: 78 (max size of a string without leading zeroes,
		// plus slack so we can copy 19 bytes every iteration).
		// We init it with zeroes, because when strconv appends the ascii representations,
		// it will omit leading zeroes.
		out     = []byte("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		divisor = newInt(10000000000000000000) // 20 digits
		y       = new(Int).set(z)              // copy to avoid modifying z
		pos     = len(out)                     // position to write to
		buf     = make([]byte, 0, 19)          // buffer to write uint64:s to
	)
	for {
		// Obtain Q and R for divisor
		var quot Int
		rem := udivrem(quot[:4], y[:4], divisor)
		y.set(&quot) // Set Q for next loop
		// Convert the R to ascii representation
		buf = strconv.AppendUint(buf[:0], rem.uint64(), 10)
		// Copy in the ascii digits
		copy(out[pos-len(buf):], buf)
		if y.isZero() {
			break
		}
		// Move 19 digits left
		pos -= 19
	}
	// skip leading zeroes by only using the 'used size' of buf
	return string(out[pos-len(buf):])
}

//////////////// MARSHALLER

// Helper function to marshal. Numeric marshals using big.Int marshaler.
func (z Int) MarshalJSON() ([]byte, error) {
	return []byte(z.String()), nil
}

// Helper function to unmarshal. Numeric unmarshals using big.Int unmarshaler.
func (z *Int) UnmarshalJSON(data []byte) error {
	val, err := z.FromString(string(data), int(defaultDecimal))
	*z = val

	return err
}
