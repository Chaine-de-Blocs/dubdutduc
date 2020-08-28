package main

import (
	"reflect"
	"testing"
)

func TestTimeSignature(t *testing.T) {
	// 0123
	hash24 := "0000123"
	hash34 := "0123"
	hash44 := "00123"
	hash54 := "000123"

	ts24 := NewTimeSignature(hash24)
	ts34 := NewTimeSignature(hash34)
	ts44 := NewTimeSignature(hash44)
	ts54 := NewTimeSignature(hash54)

	want24 := &TimeSignature{
		Numerator:   2,
		Denominator: 4,
	}
	want34 := &TimeSignature{
		Numerator:   3,
		Denominator: 4,
	}
	want44 := &TimeSignature{
		Numerator:   4,
		Denominator: 4,
	}
	want54 := &TimeSignature{
		Numerator:   5,
		Denominator: 4,
	}

	if !reflect.DeepEqual(ts24, want24) {
		t.Errorf("Time signature unmatch, want %+v has %+v", want24, ts24)
	}
	if !reflect.DeepEqual(ts34, want34) {
		t.Errorf("Time signature unmatch, want %+v has %+v", want34, ts34)
	}
	if !reflect.DeepEqual(ts44, want44) {
		t.Errorf("Time signature unmatch, want %+v has %+v", want44, ts44)
	}
	if !reflect.DeepEqual(ts54, want54) {
		t.Errorf("Time signature unmatch, want %+v has %+v", want54, ts54)
	}
}

func TestCIonianMode(t *testing.T) {
	hash := "cc00"
	m := NewMelody(hash)

	if m.Mode != Ionian {
		t.Errorf("Hash %s expected to give Ionian mode, got %d", hash, m.Mode)
	}
	if m.Scale != C {
		t.Errorf("Hash %s expected to give C scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != C {
		t.Errorf("C Ionian Tonic is expected to be C, got %d", m.Tonic())
	}
	if m.Second() != D {
		t.Errorf("C Ionian Second is expected to be D, got %d", m.Second())
	}
	if m.Third() != E {
		t.Errorf("C Ionian Third is expected to be E, got %d", m.Third())
	}
	if m.Quarte() != F {
		t.Errorf("C Ionian Quarte is expected to be F, got %d", m.Quarte())
	}
	if m.Quinte() != G {
		t.Errorf("C Ionian Quinte is expected to be G, got %d", m.Quinte())
	}
	if m.Sixte() != A {
		t.Errorf("C Ionian Sixte is expected to be A, got %d", m.Sixte())
	}
	if m.Seventh() != B {
		t.Errorf("C Ionian Seventh is expected to be B, got %d", m.Seventh())
	}
}

func TestDDorianMode(t *testing.T) {
	hash := "dd11"
	m := NewMelody(hash)

	if m.Mode != Dorian {
		t.Errorf("Hash %s expected to give Dorian mode, got %d", hash, m.Mode)
	}
	if m.Scale != D {
		t.Errorf("Hash %s expected to give D scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != D {
		t.Errorf("D Dorian Tonic is expected to be D, got %d", m.Tonic())
	}
	if m.Second() != E {
		t.Errorf("D Dorian Second is expected to be E, got %d", m.Second())
	}
	if m.Third() != F {
		t.Errorf("D Dorian Third is expected to be F, got %d", m.Third())
	}
	if m.Quarte() != G {
		t.Errorf("D Dorian Quarte is expected to be G, got %d", m.Quarte())
	}
	if m.Quinte() != A {
		t.Errorf("D Dorian Quinte is expected to be A, got %d", m.Quinte())
	}
	if m.Sixte() != B {
		t.Errorf("D Dorian Sixte is expected to be B, got %d", m.Sixte())
	}
	if m.Seventh() != C+12 {
		t.Errorf("D Dorian Seventh is expected to be C, got %d", m.Seventh())
	}
}

func TestEPhrygianMode(t *testing.T) {
	hash := "ee22"
	m := NewMelody(hash)

	if m.Mode != Phrygian {
		t.Errorf("Hash %s expected to give Phrygian mode, got %d", hash, m.Mode)
	}
	if m.Scale != E {
		t.Errorf("Hash %s expected to give E scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != E {
		t.Errorf("E Phrygian Tonic is expected to be E, got %d", m.Tonic())
	}
	if m.Second() != F {
		t.Errorf("E Phrygian Second is expected to be F, got %d", m.Second())
	}
	if m.Third() != G {
		t.Errorf("E Phrygian Third is expected to be G, got %d", m.Third())
	}
	if m.Quarte() != A {
		t.Errorf("E Phrygian Quarte is expected to be A, got %d", m.Quarte())
	}
	if m.Quinte() != B {
		t.Errorf("E Phrygian Quinte is expected to be B, got %d", m.Quinte())
	}
	if m.Sixte() != C+12 {
		t.Errorf("E Phrygian Sixte is expected to be C, got %d", m.Sixte())
	}
	if m.Seventh() != D+12 {
		t.Errorf("E Phrygian Seventh is expected to be D, got %d", m.Seventh())
	}
}

func TestFLydianMode(t *testing.T) {
	hash := "ff33"
	m := NewMelody(hash)

	if m.Mode != Lydian {
		t.Errorf("Hash %s expected to give Lydian mode, got %d", hash, m.Mode)
	}
	if m.Scale != F {
		t.Errorf("Hash %s expected to give F scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != F {
		t.Errorf("F Lydian Tonic is expected to be F, got %d", m.Tonic())
	}
	if m.Second() != G {
		t.Errorf("F Lydian Second is expected to be G, got %d", m.Second())
	}
	if m.Third() != A {
		t.Errorf("F Lydian Third is expected to be A, got %d", m.Third())
	}
	if m.Quarte() != B {
		t.Errorf("F Lydian Quarte is expected to be B, got %d", m.Quarte())
	}
	if m.Quinte() != C+12 {
		t.Errorf("F Lydian Quinte is expected to be C, got %d", m.Quinte())
	}
	if m.Sixte() != D+12 {
		t.Errorf("F Lydian Sixte is expected to be D, got %d", m.Sixte())
	}
	if m.Seventh() != E+12 {
		t.Errorf("F Lydian Seventh is expected to be E, got %d", m.Seventh())
	}
}

func TestGMixolydianMode(t *testing.T) {
	hash := "9944"
	m := NewMelody(hash)

	if m.Mode != Mixolydian {
		t.Fatalf("Hash %s expected to give Mixolydian mode, got %d", hash, m.Mode)
	}
	if m.Scale != G {
		t.Fatalf("Hash %s expected to give G scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != G {
		t.Errorf("G Mixolydian Tonic is expected to be G, got %d", m.Tonic())
	}
	if m.Second() != A {
		t.Errorf("G Mixolydian Second is expected to be A, got %d", m.Second())
	}
	if m.Third() != B {
		t.Errorf("G Mixolydian Third is expected to be B, got %d", m.Third())
	}
	if m.Quarte() != C+12 {
		t.Errorf("G Mixolydian Quarte is expected to be C, got %d", m.Quarte())
	}
	if m.Quinte() != D+12 {
		t.Errorf("G Mixolydian Quinte is expected to be D, got %d", m.Quinte())
	}
	if m.Sixte() != E+12 {
		t.Errorf("G Mixolydian Sixte is expected to be E, got %d", m.Sixte())
	}
	if m.Seventh() != F+12 {
		t.Errorf("G Mixolydian Seventh is expected to be F, got %d", m.Seventh())
	}
}

func TestAAeolianMode(t *testing.T) {
	hash := "aa55"
	m := NewMelody(hash)

	if m.Mode != Aeolian {
		t.Fatalf("Hash %s expected to give Aeolian mode, got %d", hash, m.Mode)
	}
	if m.Scale != A {
		t.Fatalf("Hash %s expected to give A scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != A {
		t.Errorf("A Aeolian Tonic is expected to be A, got %d", m.Tonic())
	}
	if m.Second() != B {
		t.Errorf("A Aeolian Second is expected to be B, got %d", m.Second())
	}
	if m.Third() != C+12 {
		t.Errorf("A Aeolian Third is expected to be C, got %d", m.Third())
	}
	if m.Quarte() != D+12 {
		t.Errorf("A Aeolian Quarte is expected to be D, got %d", m.Quarte())
	}
	if m.Quinte() != E+12 {
		t.Errorf("A Aeolian Quinte is expected to be E, got %d", m.Quinte())
	}
	if m.Sixte() != F+12 {
		t.Errorf("A Aeolian Sixte is expected to be F, got %d", m.Sixte())
	}
	if m.Seventh() != G+12 {
		t.Errorf("A Aeolian Seventh is expected to be G, got %d", m.Seventh())
	}
}
func TestBLocrianMode(t *testing.T) {
	hash := "bbb66"
	m := NewMelody(hash)

	if m.Mode != Locrian {
		t.Fatalf("Hash %s expected to give Locrian mode, got %d", hash, m.Mode)
	}
	if m.Scale != B {
		t.Fatalf("Hash %s expected to give B scale, got %d", hash, m.Scale)
	}

	if m.Tonic() != B {
		t.Errorf("B Locrian Tonic is expected to be B, got %d", m.Tonic())
	}
	if m.Second() != C+12 {
		t.Errorf("B Locrian Second is expected to be C, got %d", m.Second())
	}
	if m.Third() != D+12 {
		t.Errorf("B Locrian Third is expected to be D, got %d", m.Third())
	}
	if m.Quarte() != E+12 {
		t.Errorf("B Locrian Quarte is expected to be E, got %d", m.Quarte())
	}
	if m.Quinte() != F+12 {
		t.Errorf("B Locrian Quinte is expected to be F, got %d", m.Quinte())
	}
	if m.Sixte() != G+12 {
		t.Errorf("B Locrian Sixte is expected to be G, got %d", m.Sixte())
	}
	if m.Seventh() != A+12 {
		t.Errorf("B Locrian Seventh is expected to be A, got %d", m.Seventh())
	}
}
