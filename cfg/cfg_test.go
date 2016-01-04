package cfg

import (
	"testing"
)

func TestGetCfg(t *testing.T) {
	obtainedCfg, obtainedErr := GetCfg()

	if obtainedCfg == nil {
		t.Error("Expected *Cfg, got nil")
	}

	if obtainedErr != nil {
		t.Errorf("Expected nil error. Got: %s", obtainedErr)
	}
}

func TestInArgsValidateSuccess(t *testing.T) {
	testCases := []struct {
		inArgs *InArgs
		hasErr bool
	}{
		{
			inArgs: &InArgs{Bucket: "test-bucket-name"},
			hasErr: false,
		},
		{
			inArgs: &InArgs{},
			hasErr: true,
		},
	}

	for i, testCase := range testCases {
		obtainerErr := testCase.inArgs.Validate()
		if (obtainerErr != nil) != testCase.hasErr {
			t.Errorf("Test case: %d. Expected hasError: %t. Got error: %s", i, testCase.hasErr, obtainerErr)
		}
	}
}
