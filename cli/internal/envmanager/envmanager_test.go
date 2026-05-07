package envmanager

import "testing"

func TestEncryptDecryptVariablesRoundtrip(t *testing.T) {
	in := map[string]string{"A": "1", "B": "hello"}
	enc, err := EncryptVariables(in, "pass")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	out, err := DecryptVariables(enc, "pass")
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if out["A"] != "1" || out["B"] != "hello" {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestDecryptVariables_WrongPasswordFails(t *testing.T) {
	in := map[string]string{"A": "1"}
	enc, err := EncryptVariables(in, "pass")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	if _, err := DecryptVariables(enc, "wrong"); err == nil {
		t.Fatalf("expected decrypt error")
	}
}

func TestDecryptVariables_TamperedCiphertextFails(t *testing.T) {
	in := map[string]string{"A": "secret"}
	enc, err := EncryptVariables(in, "pass")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	// Tamper with the encrypted value
	tampered := make(map[string]string)
	for k, v := range enc {
		// Flip a character in the middle of the base64 blob
		if len(v) > 10 {
			b := []byte(v)
			b[len(b)/2] ^= 0xFF
			tampered[k] = string(b)
		} else {
			tampered[k] = v
		}
	}
	if _, err := DecryptVariables(tampered, "pass"); err == nil {
		t.Fatalf("expected tampered ciphertext to fail decryption")
	}
}

func TestMergeVariables(t *testing.T) {
	left := map[string]string{"A": "1", "B": "2"}
	right := map[string]string{"B": "3", "C": "4"}
	out := MergeVariables(left, right)
	if out["A"] != "1" || out["B"] != "3" || out["C"] != "4" {
		t.Fatalf("unexpected merge: %#v", out)
	}
}
