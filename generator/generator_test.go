package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGRPCCodeGenInfoValidateNoError(t *testing.T) {
	assert := assert.New(t)
	cases := []GRPCCodeGenInfo{
		{
			"package",
			"ServiceName",
			[]GRPCMethod{
				{
					"Method1",
					"Request",
					"Response",
				},
				{
					"Method2",
					"Request",
					"Response",
				},
			},
		},
	}
	for _, c := range cases {
		err := c.Validate()
		assert.NoError(err)
	}
}

func TestGRPCCodeGenInfoValidateError(t *testing.T) {
	assert := assert.New(t)
	cases := []GRPCCodeGenInfo{
		{
			"",
			"ServiceName",
			[]GRPCMethod{
				{
					"Method1",
					"Request",
					"Response",
				},
				{
					"Method2",
					"Request",
					"Response",
				},
			},
		},
		{
			"package",
			"",
			[]GRPCMethod{
				{
					"Method1",
					"Request",
					"Response",
				},
				{
					"Method2",
					"Request",
					"Response",
				},
			},
		},
		{
			"package",
			"ServiceName",
			[]GRPCMethod{},
		},
		{
			"package",
			"ServiceName",
			[]GRPCMethod{
				{
					"",
					"Request",
					"Response",
				},
				{
					"Method2",
					"Request",
					"Response",
				},
			},
		},
		{
			"package",
			"ServiceName",
			[]GRPCMethod{
				{
					"Method1",
					"Request",
					"Response",
				},
				{
					"Method2",
					"",
					"Response",
				},
			},
		},
		{
			"package",
			"ServiceName",
			[]GRPCMethod{
				{
					"Method1",
					"Request",
					"Response",
				},
				{
					"Method2",
					"Request",
					"",
				},
			},
		},
	}
	for _, c := range cases {
		err := c.Validate()
		assert.Error(err)
	}
}

func TestGenerateGRPCTestCode(t *testing.T) {
	assert := assert.New(t)
	grpcCodeGenInfo := GRPCCodeGenInfo{
		Package:         "pb",
		GRPCServiceName: "TestService",
		GRPCMethods: []GRPCMethod{
			{
				Name:         "Hello",
				RequestType:  "HReq",
				ResponseType: "HRes",
			},
			{
				Name:         "Bye",
				RequestType:  "BReq",
				ResponseType: "BRes",
			},
		},
	}
	code, err := GenerateGRPCTestCode(grpcCodeGenInfo)
	assert.Equal(expectedCode, code)
	assert.NoError(err)
}

var expectedCode = `
package pb

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// TestServiceTestRunner is a runner to run the TestService service test.
type TestServiceTestRunner struct {
	Client TestServiceClient
}

// NewTestClient returns new TestServiceRunner.
func NewTestClient(client TestServiceClient) *TestServiceTestRunner {
	return &TestServiceTestRunner{
		Client: client,
	}
}

// RunGRPCTest sends a gPRC request according to the scenario written in the JSON file and tests the response.
// testHandlerMap takes a gRPC method name as a key and value has a function that compares expected response and actual response and defines how to handle the test.
func (runner *TestServiceTestRunner) RunGRPCTest(t *testing.T, jsonPath string, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
	scenarioData, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		panic(err)
	}
	var scenario []map[string]interface{}
	json.Unmarshal(scenarioData, &scenario)
	for _, testCase := range scenario {
		ctx := context.Background()
		runner.runTest(ctx, t, testCase, testHandlerMap)
	}
}

const (
	actionJSONKey            = "action"
	requestJSONKey           = "request"
	expectedResponseJSONKey  = "expected_response"
	errorExpectationJSONKey  = "error_expectation"
	expectedErrorCodeJSONKey = "expected_error_code"
	skipMessage              = "Skip comparing expected response and actual response"
)

func (runner *TestServiceTestRunner) runTest(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandlerMap map[string]*func(t *testing.T, expectedResponse, response interface{})) {
	action := testCase[actionJSONKey].(string)
	f := func(t *testing.T) {
		switch action {
		case "Hello":
			testHandler := testHandlerMap["Hello"]
			runner.testHello(ctx, t, testCase, testHandler)
		case "Bye":
			testHandler := testHandlerMap["Bye"]
			runner.testBye(ctx, t, testCase, testHandler)
		}
	}
	t.Run(action, f)
}

func (runner *TestServiceTestRunner) testHello(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandler *func(t *testing.T, expectedResponse, response interface{})) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := HReq{}
	json.Unmarshal(reqJSON, &req)
	res, err := runner.Client.Hello(ctx, &req)
	errExpectation := testCase[errorExpectationJSONKey].(bool)
	if errExpectation {
		errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
		errCodeU := uint32(errCodeF)
		expectedErrCode := codes.Code(errCodeU)
		if expectedErrCode != grpc.Code(err) {
			t.Fatalf("The error code of the response of Hello is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
		}
	} else {
		resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
		if resErr != nil {
			panic(resErr)
		}
		expectedRes := HRes{}
		json.Unmarshal(resJSON, &expectedRes)
		if testHandler != nil {
			handler := *testHandler
			handler(t, expectedRes, *res)
		} else {
			if !reflect.DeepEqual(expectedRes, *res) {
				t.Fatal("The actual response of the Hello was not equal to the expected response.")
			}
		}
	}
}

func (runner *TestServiceTestRunner) testBye(ctx context.Context, t *testing.T, testCase map[string]interface{}, testHandler *func(t *testing.T, expectedResponse, response interface{})) {
	reqJSON, reqErr := json.Marshal(testCase[requestJSONKey])
	if reqErr != nil {
		panic(reqErr)
	}
	req := BReq{}
	json.Unmarshal(reqJSON, &req)
	res, err := runner.Client.Bye(ctx, &req)
	errExpectation := testCase[errorExpectationJSONKey].(bool)
	if errExpectation {
		errCodeF := testCase[expectedErrorCodeJSONKey].(float64)
		errCodeU := uint32(errCodeF)
		expectedErrCode := codes.Code(errCodeU)
		if expectedErrCode != grpc.Code(err) {
			t.Fatalf("The error code of the response of Bye is not as expected. Expected: %d, Actual: %d\n", expectedErrCode, grpc.Code(err))
		}
	} else {
		resJSON, resErr := json.Marshal(testCase[expectedResponseJSONKey])
		if resErr != nil {
			panic(resErr)
		}
		expectedRes := BRes{}
		json.Unmarshal(resJSON, &expectedRes)
		if testHandler != nil {
			handler := *testHandler
			handler(t, expectedRes, *res)
		} else {
			if !reflect.DeepEqual(expectedRes, *res) {
				t.Fatal("The actual response of the Bye was not equal to the expected response.")
			}
		}
	}
}

`