package tokens

import (
	"testing"
)

func TestWrappers_DelegateToDefaultGenerators(t *testing.T) {
	origJWT := DefaultJwtGenerator
	origOpaque := DefaultOpaqueGenerator
	t.Cleanup(func() {
		DefaultJwtGenerator = origJWT
		DefaultOpaqueGenerator = origOpaque
	})

	jwtMock := NewMockGenerator()
	retTok := "jwt-token"
	jwtMock.On("Generate", "sub", Principal{"p": 1}).Return(&retTok, nil)
	jwtMock.On("Validate", "jwt-token").Return(Principal{"ok": true}, nil)
	DefaultJwtGenerator = jwtMock

	opMock := NewMockGenerator()
	retTok2 := "op-token"
	opMock.On("Generate", "sub2", Principal{"q": 2}).Return(&retTok2, nil)
	opMock.On("Validate", "op-token").Return(Principal{"ok2": true}, nil)
	DefaultOpaqueGenerator = opMock

	t1, err := JwtGenerate("sub", Principal{"p": 1})
	if err != nil || t1 == nil || *t1 != "jwt-token" {
		t.Fatalf("unexpected jwt wrapper result: %v %v", t1, err)
	}
	p1, err := JwtValidate("jwt-token")
	if err != nil || p1["ok"].(bool) != true {
		t.Fatalf("unexpected jwt validate result: %v %v", p1, err)
	}

	t2, err := OpaqueGenerate("sub2", Principal{"q": 2})
	if err != nil || t2 == nil || *t2 != "op-token" {
		t.Fatalf("unexpected opaque wrapper result: %v %v", t2, err)
	}
	p2, err := OpaqueValidate("op-token")
	if err != nil || p2["ok2"].(bool) != true {
		t.Fatalf("unexpected opaque validate result: %v %v", p2, err)
	}

	jwtMock.AssertExpectations(t)
	opMock.AssertExpectations(t)
}
