package requestid

import (
	"context"
	"testing"

	"go.unistack.org/micro/v4/metadata"
)

func TestDefaultMetadataFunc(t *testing.T) {
	ctx := context.TODO()
	var err error

	ctx, err = DefaultMetadataFunc(ctx)
	if err != nil {
		t.Fatalf("%v", err)
	}

	imd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		t.Fatalf("md missing in incoming context")
	}
	omd, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatalf("md missing in outgoing context")
	}

	iv, iok := imd.Get(DefaultMetadataKey)
	ov, ook := omd.Get(DefaultMetadataKey)

	if !iok || !ook {
		t.Fatalf("missing metadata key value")
	}
	if iv != ov {
		t.Fatalf("invalid metadata key value")
	}
}
