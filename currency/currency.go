package currency

import "math"

// Money : a structure type for money
type Money struct {
	Integer    int64
	Fractional int64
}

const fractionPrecision = 100000000000000.0
const fractionDisplayPrecision = fractionPrecision / 10.0

// SetVal set initial value to Money
func (m *Money) SetVal(val float64) {
	integer, frac := math.Modf(val)
	m.Integer = int64(integer)
	m.Fractional = int64(frac * fractionPrecision)
}

// MoneyCtr "constructor"
func MoneyCtr(val float64) Money {
	var m Money
	m.SetVal(val)
	return m
}

// GetVal : get value as float64
func (m *Money) GetVal() float64 {
	intermediateResult := float64(m.Integer) + float64(m.Fractional)/fractionPrecision
	roundedVal := math.Round(intermediateResult * fractionDisplayPrecision)
	return roundedVal / fractionDisplayPrecision
}

// ToggleSign : just multiply by -1
func (m *Money) ToggleSign() {
	m.Integer = -m.Integer
	m.Fractional = -m.Fractional
}

// Add : add value
func (m *Money) Add(addVal Money) {
	m.Integer = m.Integer + addVal.Integer
	frac := m.Fractional + addVal.Fractional
	if frac > 0 && frac >= int64(fractionPrecision) {
		m.Integer++
		frac = frac % int64(fractionPrecision)
	}
	if frac < 0 && frac <= -int64(fractionPrecision) {
		m.Integer--
		frac = frac % int64(fractionPrecision)
	}
	m.Fractional = frac
}

// Sub : subtract value
func (m *Money) Sub(subtractVal Money) {
	subtractVal.ToggleSign()
	m.Add(subtractVal)
}

// MultiplyBy : multiply by some value
func (m *Money) MultiplyBy(mul float64) {
	m.SetVal(m.GetVal() * mul)
}

// DivideBy : divide by some value
func (m *Money) DivideBy(div float64) {
	m.SetVal(m.GetVal() / div)
}

func isAlmostEqual(valueOne, valueTwo int64) bool {
	if valueOne-valueTwo >= 0 {
		return valueOne-valueTwo < 10
	}
	return valueTwo-valueOne < 10
}

// IsEqualTo check for equality
func (m *Money) IsEqualTo(val Money) bool {
	if m.Integer == val.Integer && isAlmostEqual(m.Fractional, val.Fractional) {
		return true
	}
	return false
}

// IsGreaterThan isGreater comparison
func (m *Money) IsGreaterThan(val Money) bool {
	if m.IsEqualTo(val) {
		return false
	}
	if m.Integer > val.Integer {
		return true
	} else if m.Integer == val.Integer && m.Fractional > val.Fractional {
		return true
	}
	return false
}

// IsLessThan isLess comparison
func (m *Money) IsLessThan(val Money) bool {
	if m.IsEqualTo(val) || m.IsGreaterThan(val) {
		return false
	}
	return true
}
