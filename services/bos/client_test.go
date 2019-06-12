package bos

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
	"github.com/baidubce/bce-sdk-go/util/log"
)

var (
	BOS_CLIENT    *Client
	EXISTS_BUCKET = "bos-rd-ssy"
)

// For security reason, ak/sk should not hard write here.
type Conf struct {
	AK string
	SK string
}

func init() {
	_, f, _, _ := runtime.Caller(0)
	for i := 0; i < 7; i++ {
		f = filepath.Dir(f)
	}
	conf := filepath.Join(f, "config.json")
	fp, err := os.Open(conf)
	if err != nil {
		log.Fatal("config json file of ak/sk not given:", conf)
		os.Exit(1)
	}
	decoder := json.NewDecoder(fp)
	confObj := &Conf{}
	decoder.Decode(confObj)

	BOS_CLIENT, _ = NewClient(confObj.AK, confObj.SK, "")
	//log.SetLogHandler(log.STDERR | log.FILE)
	//log.SetRotateType(log.ROTATE_SIZE)
	log.SetLogLevel(log.WARN)

	//log.SetLogHandler(log.STDERR)
	//log.SetLogLevel(log.DEBUG)
}

// ExpectEqual is the helper function for test each case
func ExpectEqual(alert func(format string, args ...interface{}),
	expected interface{}, actual interface{}) bool {
	expectedValue, actualValue := reflect.ValueOf(expected), reflect.ValueOf(actual)
	equal := false
	switch {
	case expected == nil && actual == nil:
		return true
	case expected != nil && actual == nil:
		equal = expectedValue.IsNil()
	case expected == nil && actual != nil:
		equal = actualValue.IsNil()
	default:
		if actualType := reflect.TypeOf(actual); actualType != nil {
			if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
				equal = reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
			}
		}
	}
	if !equal {
		_, file, line, _ := runtime.Caller(1)
		alert("%s:%d: missmatch, expect %v but %v", file, line, expected, actual)
		return false
	}
	return true
}

func TestListBuckets(t *testing.T) {
	res, err := BOS_CLIENT.ListBuckets()
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestListObjects(t *testing.T) {
	args := &api.ListObjectsArgs{Prefix: "test", MaxKeys: 10}
	res, err := BOS_CLIENT.ListObjects(EXISTS_BUCKET, args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestSimpleListObjects(t *testing.T) {
	res, err := BOS_CLIENT.SimpleListObjects(EXISTS_BUCKET, "test", 10, "", "")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestHeadBucket(t *testing.T) {
	err := BOS_CLIENT.HeadBucket(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
}

func TestDoesBucketExist(t *testing.T) {
	exist, err := BOS_CLIENT.DoesBucketExist(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, exist, true)
	ExpectEqual(t.Errorf, err, nil)

	exist, err = BOS_CLIENT.DoesBucketExist("xxx")
	ExpectEqual(t.Errorf, exist, false)
}

func TestPutBucket(t *testing.T) {
	res, err := BOS_CLIENT.PutBucket("test-put-bucket")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
}

func TestDeleteBucket(t *testing.T) {
	err := BOS_CLIENT.DeleteBucket("test-put-bucket")
	ExpectEqual(t.Errorf, err, nil)
}

func TestGetBucketLocation(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketLocation(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
}

func TestPutBucketAcl(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[{
                "id":"e13b12d0131b4c8bae959df4969387b8"
            }],
            "permission":["FULL_CONTROL"]
        }
    ]
}`
	body, _ := bce.NewBodyFromString(acl)
	err := BOS_CLIENT.PutBucketAcl(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketAcl(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[0].Id,
		"e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutBucketAclFromCanned(t *testing.T) {
	err := BOS_CLIENT.PutBucketAclFromCanned(EXISTS_BUCKET, api.CANNED_ACL_PUBLIC_READ)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketAcl(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[0].Id, "*")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "READ")
}

func TestPutBucketAclFromFile(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[
                {"id":"e13b12d0131b4c8bae959df4969387b8"},
                {"id":"a13b12d0131b4c8bae959df4969387b8"}
            ],
            "permission":["FULL_CONTROL"]
        }
    ]
}`
	fname := "/tmp/test-put-bucket-acl-by-file"
	f, _ := os.Create(fname)
	f.WriteString(acl)
	f.Close()
	err := BOS_CLIENT.PutBucketAclFromFile(EXISTS_BUCKET, fname)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketAcl(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	os.Remove(fname)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[0].Id,
		"e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[1].Id,
		"a13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutBucketAclFromString(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[
                {"id":"e13b12d0131b4c8bae959df4969387b8"},
                {"id":"a13b12d0131b4c8bae959df4969387b8"}
            ],
            "permission":["FULL_CONTROL"]
        }
    ]
}`
	err := BOS_CLIENT.PutBucketAclFromString(EXISTS_BUCKET, acl)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketAcl(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[0].Id,
		"e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[1].Id,
		"a13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutBucketAclFromStruct(t *testing.T) {
	args := &api.PutBucketAclArgs{
		[]api.GrantType{
			api.GrantType{
				Grantee: []api.GranteeType{
					api.GranteeType{"e13b12d0131b4c8bae959df4969387b8"},
				},
				Permission: []string{
					"FULL_CONTROL",
				},
			},
		},
	}
	err := BOS_CLIENT.PutBucketAclFromStruct(EXISTS_BUCKET, args)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketAcl(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Grantee[0].Id,
		"e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutBucketLogging(t *testing.T) {
	body, _ := bce.NewBodyFromString(
		`{"targetBucket": "bos-rd-ssy", "targetPrefix": "my-log/"}`)
	err := BOS_CLIENT.PutBucketLogging(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLogging(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.TargetBucket, "bos-rd-ssy")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.TargetPrefix, "my-log/")
}

func TestPutBucketLoggingFromString(t *testing.T) {
	logging := `{"targetBucket": "bos-rd-ssy", "targetPrefix": "my-log2/"}`
	err := BOS_CLIENT.PutBucketLoggingFromString(EXISTS_BUCKET, logging)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLogging(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.TargetBucket, "bos-rd-ssy")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.TargetPrefix, "my-log2/")
}

func TestPutBucketLoggingFromStruct(t *testing.T) {
	obj := &api.PutBucketLoggingArgs{"bos-rd-ssy", "my-log3/"}
	err := BOS_CLIENT.PutBucketLoggingFromStruct(EXISTS_BUCKET, obj)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLogging(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.TargetBucket, "bos-rd-ssy")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.TargetPrefix, "my-log3/")
}

func TestDeleteBucketLogging(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketLogging(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLogging(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Status, "disabled")
}

func TestPutBucketLifecycle(t *testing.T) {
	str := `{
    "rule": [
        {
            "id": "transition-to-cold",
            "status": "enabled",
            "resource": ["bos-rd-ssy/test*"],
            "condition": {
                "time": {
                    "dateGreaterThan": "2018-09-07T00:00:00Z"
                }
            },
            "action": {
                "name": "DeleteObject"
            }
        }
    ]
}`
	body, _ := bce.NewBodyFromString(str)
	err := BOS_CLIENT.PutBucketLifecycle(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLifecycle(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Rule[0].Id, "transition-to-cold")
	ExpectEqual(t.Errorf, res.Rule[0].Status, "enabled")
	ExpectEqual(t.Errorf, res.Rule[0].Resource[0], "bos-rd-ssy/test*")
	ExpectEqual(t.Errorf, res.Rule[0].Condition.Time.DateGreaterThan, "2018-09-07T00:00:00Z")
	ExpectEqual(t.Errorf, res.Rule[0].Action.Name, "DeleteObject")
}

func TestPutBucketLifecycleFromString(t *testing.T) {
	obj := `{
    "rule": [
        {
            "id": "transition-to-cold",
            "status": "enabled",
            "resource": ["bos-rd-ssy/test*"],
            "condition": {
                "time": {
                    "dateGreaterThan": "2018-09-07T00:00:00Z"
                }
            },
            "action": {
                "name": "DeleteObject"
            }
        }
    ]
}`
	err := BOS_CLIENT.PutBucketLifecycleFromString(EXISTS_BUCKET, obj)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketLifecycle(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Rule[0].Id, "transition-to-cold")
	ExpectEqual(t.Errorf, res.Rule[0].Status, "enabled")
	ExpectEqual(t.Errorf, res.Rule[0].Resource[0], "bos-rd-ssy/test*")
	ExpectEqual(t.Errorf, res.Rule[0].Condition.Time.DateGreaterThan, "2018-09-07T00:00:00Z")
	ExpectEqual(t.Errorf, res.Rule[0].Action.Name, "DeleteObject")
}

func TestDeleteBucketLifecycle(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketLifecycle(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, _ := BOS_CLIENT.GetBucketLifecycle(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, res, nil)
}

func TestPutBucketStorageClass(t *testing.T) {
	err := BOS_CLIENT.PutBucketStorageclass(EXISTS_BUCKET, api.STORAGE_CLASS_STANDARD_IA)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStorageclass(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res, api.STORAGE_CLASS_STANDARD_IA)
}

func TestGetBucketStorageClass(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketStorageclass(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestPutBucketReplication(t *testing.T) {
	BOS_CLIENT.DeleteBucketReplication(EXISTS_BUCKET)
	str := `{
		"id": "abc",
		"status":"enabled",
		"resource": ["bos-rd-ssy/films"],
		"destination": {
			"bucket": "bos-rd-su-test",
			"storageClass": "COLD"
		},
		"replicateDeletes": "disabled"
}`
	body, _ := bce.NewBodyFromString(str)
	err := BOS_CLIENT.PutBucketReplication(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Id, "abc")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.Resource[0], "bos-rd-ssy/films")
	ExpectEqual(t.Errorf, res.Destination.Bucket, "bos-rd-su-test")
	ExpectEqual(t.Errorf, res.ReplicateDeletes, "disabled")
}

func TestPutBucketReplicationFromFile(t *testing.T) {
	BOS_CLIENT.DeleteBucketReplication(EXISTS_BUCKET)
	str := `{
		"id": "abc",
		"status":"enabled",
		"resource": ["bos-rd-ssy/films"],
		"destination": {
			"bucket": "bos-rd-su-test",
			"storageClass": "COLD"
		},
		"replicateDeletes": "disabled"
}`
	fname := "/tmp/test-put-bucket-replication-by-file"
	f, _ := os.Create(fname)
	f.WriteString(str)
	f.Close()
	err := BOS_CLIENT.PutBucketReplicationFromFile(EXISTS_BUCKET, fname)
	os.Remove(fname)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Id, "abc")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.Resource[0], "bos-rd-ssy/films")
	ExpectEqual(t.Errorf, res.Destination.Bucket, "bos-rd-su-test")
	ExpectEqual(t.Errorf, res.ReplicateDeletes, "disabled")
}

func TestPutBucketReplicationFromString(t *testing.T) {
	BOS_CLIENT.DeleteBucketReplication(EXISTS_BUCKET)
	str := `{
		"id": "abc",
		"status":"enabled",
		"resource": ["bos-rd-ssy/films"],
		"destination": {
			"bucket": "bos-rd-su-test",
			"storageClass": "COLD"
		},
		"replicateDeletes": "disabled"
}`
	err := BOS_CLIENT.PutBucketReplicationFromString(EXISTS_BUCKET, str)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Id, "abc")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.Resource[0], "bos-rd-ssy/films")
	ExpectEqual(t.Errorf, res.Destination.Bucket, "bos-rd-su-test")
	ExpectEqual(t.Errorf, res.ReplicateDeletes, "disabled")
}

func TestPutBucketReplicationFromStruct(t *testing.T) {
	BOS_CLIENT.DeleteBucketReplication(EXISTS_BUCKET)
	args := &api.PutBucketReplicationArgs{
		Id:               "abc",
		Status:           "enabled",
		Resource:         []string{"bos-rd-ssy/films"},
		Destination:      &api.BucketReplicationDescriptor{"bos-rd-su-test", "COLD"},
		ReplicateDeletes: "disabled",
	}
	err := BOS_CLIENT.PutBucketReplicationFromStruct(EXISTS_BUCKET, args)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Id, "abc")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.Resource[0], "bos-rd-ssy/films")
	ExpectEqual(t.Errorf, res.Destination.Bucket, "bos-rd-su-test")
	ExpectEqual(t.Errorf, res.ReplicateDeletes, "disabled")
}

func TestGetBucketReplication(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Id, "abc")
	ExpectEqual(t.Errorf, res.Status, "enabled")
	ExpectEqual(t.Errorf, res.Resource[0], "bos-rd-ssy/films")
	ExpectEqual(t.Errorf, res.Destination.Bucket, "bos-rd-su-test")
	ExpectEqual(t.Errorf, res.ReplicateDeletes, "disabled")
}

func TestGetBucketReplicationProcess(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketReplicationProgress(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
}

func TestDeleteBucketReplication(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketReplication(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
}

func TestPutBucketEncryption(t *testing.T) {
	err := BOS_CLIENT.PutBucketEncryption(EXISTS_BUCKET, api.ENCRYPTION_AES256)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketEncryption(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res, api.ENCRYPTION_AES256)
}

func TestGetBucketEncryption(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketEncryption(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestDeleteBucketEncryption(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketEncryption(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketEncryption(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestPutBucketStaticWebsite(t *testing.T) {
	BOS_CLIENT.DeleteBucketStaticWebsite(EXISTS_BUCKET)
	body, _ := bce.NewBodyFromString(`{"index": "index.html", "notFound":"blank.html"}`)
	err := BOS_CLIENT.PutBucketStaticWebsite(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Index, "index.html")
	ExpectEqual(t.Errorf, res.NotFound, "blank.html")
}

func TestPutBucketStaticWebsiteFromString(t *testing.T) {
	BOS_CLIENT.DeleteBucketStaticWebsite(EXISTS_BUCKET)
	jsonConf := `{"index": "index.html", "notFound":"blank.html"}`
	err := BOS_CLIENT.PutBucketStaticWebsiteFromString(EXISTS_BUCKET, jsonConf)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Index, "index.html")
	ExpectEqual(t.Errorf, res.NotFound, "blank.html")
}

func TestPutBucketStaticWebsiteFromStruct(t *testing.T) {
	BOS_CLIENT.DeleteBucketStaticWebsite(EXISTS_BUCKET)
	obj := &api.PutBucketStaticWebsiteArgs{"index.html", "blank.html"}
	err := BOS_CLIENT.PutBucketStaticWebsiteFromStruct(EXISTS_BUCKET, obj)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Index, "index.html")
	ExpectEqual(t.Errorf, res.NotFound, "blank.html")
}

func TestSimplePutBucketStaticWebsite(t *testing.T) {
	BOS_CLIENT.DeleteBucketStaticWebsite(EXISTS_BUCKET)
	err := BOS_CLIENT.SimplePutBucketStaticWebsite(EXISTS_BUCKET, "index.html", "blank.html")
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.Index, "index.html")
	ExpectEqual(t.Errorf, res.NotFound, "blank.html")
}

func TestGetBucketStaticWebsite(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
}

func TestDeleteBucketStaticWebsite(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketStaticWebsite(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v", res)
}

func TestPutBucketCors(t *testing.T) {
	body, _ := bce.NewBodyFromString(`
	{
		"corsConfiguration": [
			{
				"allowedOrigins": ["https://www.baidu.com"],
				"allowedMethods": ["GET"],
				"maxAgeSeconds": 1800
			}
		]
	}
	`)
	err := BOS_CLIENT.PutBucketCors(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedOrigins[0], "https://www.baidu.com")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedMethods[0], "GET")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].MaxAgeSeconds, 1800)
}

func TestPutBucketCorsFromFile(t *testing.T) {
	str := `{
		"corsConfiguration": [
			{
				"allowedOrigins": ["https://www.baidu.com"],
				"allowedMethods": ["GET"],
				"maxAgeSeconds": 1800
			}
		]
	}`
	fname := "/tmp/test-put-bucket-cors-by-file"
	f, _ := os.Create(fname)
	f.WriteString(str)
	f.Close()
	err := BOS_CLIENT.PutBucketCorsFromFile(EXISTS_BUCKET, fname)
	os.Remove(fname)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedOrigins[0], "https://www.baidu.com")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedMethods[0], "GET")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].MaxAgeSeconds, 1800)

	err = BOS_CLIENT.PutBucketCorsFromFile(EXISTS_BUCKET, "/tmp/not-exist")
	ExpectEqual(t.Errorf, err != nil, true)
}

func TestPutBucketCorsFromString(t *testing.T) {
	str := `{
		"corsConfiguration": [
			{
				"allowedOrigins": ["https://www.baidu.com"],
				"allowedMethods": ["GET"],
				"maxAgeSeconds": 1800
			}
		]
	}`
	err := BOS_CLIENT.PutBucketCorsFromString(EXISTS_BUCKET, str)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedOrigins[0], "https://www.baidu.com")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedMethods[0], "GET")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].MaxAgeSeconds, 1800)

	err = BOS_CLIENT.PutBucketCorsFromString(EXISTS_BUCKET, "")
	ExpectEqual(t.Errorf, err != nil, true)
}

func TestPutBucketCorsFromStruct(t *testing.T) {
	obj := &api.PutBucketCorsArgs{
		[]api.BucketCORSType{
			api.BucketCORSType{
				AllowedOrigins: []string{"https://www.baidu.com"},
				AllowedMethods: []string{"GET"},
				MaxAgeSeconds:  1200,
			},
		},
	}
	err := BOS_CLIENT.PutBucketCorsFromStruct(EXISTS_BUCKET, obj)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedOrigins[0], "https://www.baidu.com")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedMethods[0], "GET")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].MaxAgeSeconds, 1200)

	err = BOS_CLIENT.PutBucketCorsFromStruct(EXISTS_BUCKET, nil)
	ExpectEqual(t.Errorf, err != nil, true)
}

func TestGetBucketCors(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedOrigins[0], "https://www.baidu.com")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].AllowedMethods[0], "GET")
	ExpectEqual(t.Errorf, res.CorsConfiguration[0].MaxAgeSeconds, 1200)
}

func TestDeleteBucketCors(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCors(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v, %v", res, err)
}

func TestPutBucketCopyrightProtection(t *testing.T) {
	err := BOS_CLIENT.PutBucketCopyrightProtection(EXISTS_BUCKET,
		"bos-rd-ssy/glog.go", "bos-rd-ssy/films/*")
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCopyrightProtection(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, res[0], "bos-rd-ssy/glog.go")
	ExpectEqual(t.Errorf, res[1], "bos-rd-ssy/films/*")
}

func TestGetBucketCopyrightProtection(t *testing.T) {
	res, err := BOS_CLIENT.GetBucketCopyrightProtection(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v, %v", res, err)
}

func TestDeleteBucketCopyrightProtection(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketCopyrightProtection(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketCopyrightProtection(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v, %v", res, err)
}

func TestPutBucketNotification(t *testing.T) {
	body, _ := bce.NewBodyFromString(`
	{
		"notifications": [
			{
				"id": "id-1",
				"appId": "app-id-1",
				"status": "enabled",
				"resources": [
					"/path1/*", "/*.jpg", "/*.png"
				],
				"events": [
					"PutObject"
				],
				"quota": {
					"quotaDay": 1000,
					"quotaSec": 10
				},
				"apps": [
					{
						"id": "app-id-1",
						"eventUrl": "http://127.0.0.1:8080/event"
					}
				]
			}
		]
	}`)
	err := BOS_CLIENT.PutBucketNotification(EXISTS_BUCKET, body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketNotification(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, len(res.Notifications), 1)
	notification := res.Notifications[0]
	ExpectEqual(t.Errorf, notification.Id, "id-1")
	ExpectEqual(t.Errorf, notification.AppId, "app-id-1")
	ExpectEqual(t.Errorf, notification.Status, "enabled")
	ExpectEqual(t.Errorf, notification.Resources, []string{"/path1/*", "/*.jpg", "/*.png"})
	ExpectEqual(t.Errorf, notification.Events, []string{"PutObject"})
	ExpectEqual(t.Errorf, notification.Quota.QuotaDay, float64(1000))
	ExpectEqual(t.Errorf, notification.Quota.QuotaSec, float64(10))
	ExpectEqual(t.Errorf, notification.Apps[0].Id, "app-id-1")
	ExpectEqual(t.Errorf, notification.Apps[0].EventUrl, "http://127.0.0.1:8080/event")
}

func TestPutBucketNotificationFromFile(t *testing.T) {
	str := `
	{
		"notifications": [
			{
				"id": "id-1",
				"appId": "app-id-1",
				"status": "enabled",
				"resources": [
					"/path1/*", "/*.jpg", "/*.png"
				],
				"events": [
					"PutObject"
				],
				"quota": {
					"quotaDay": 1000,
					"quotaSec": 10
				},
				"apps": [
					{
						"id": "app-id-1",
						"eventUrl": "http://127.0.0.1:8080/event"
					}
				]
			}
		]
	}`
	filename := "/tmp/test-put-notification-by-file"
	f, _ := os.Create(filename)
	defer os.Remove(filename)
	f.WriteString(str)
	f.Close()

	err := BOS_CLIENT.PutBucketNotificationFromFile(EXISTS_BUCKET, filename)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketNotification(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, len(res.Notifications), 1)
	notification := res.Notifications[0]
	ExpectEqual(t.Errorf, notification.Id, "id-1")
	ExpectEqual(t.Errorf, notification.AppId, "app-id-1")
	ExpectEqual(t.Errorf, notification.Status, "enabled")
	ExpectEqual(t.Errorf, notification.Resources, []string{"/path1/*", "/*.jpg", "/*.png"})
	ExpectEqual(t.Errorf, notification.Events, []string{"PutObject"})
	ExpectEqual(t.Errorf, notification.Quota.QuotaDay, float64(1000))
	ExpectEqual(t.Errorf, notification.Quota.QuotaSec, float64(10))
	ExpectEqual(t.Errorf, notification.Apps[0].Id, "app-id-1")
	ExpectEqual(t.Errorf, notification.Apps[0].EventUrl, "http://127.0.0.1:8080/event")

	err = BOS_CLIENT.PutBucketNotificationFromFile(EXISTS_BUCKET, "/tmp/not-exist")
	ExpectEqual(t.Errorf, err != nil, true)
}

func TestPutBucketNotificationFromStruct(t *testing.T) {
	args := &api.PutBucketNotificationArgs{
		Notifications: []api.Notification{
			api.Notification{
				Id:        "id-1",
				AppId:     "app-id-1",
				Status:    "enabled",
				Resources: []string{"/path1/*", "/*.jpg", "/*.png"},
				Events:    []string{"PutObject"},
				Quota: api.NotificationQuota{
					QuotaDay: 1000,
					QuotaSec: 10,
				},
				Apps: []api.NotificationApp{
					api.NotificationApp{
						Id:       "app-id-1",
						EventUrl: "http://127.0.0.1:8080/event",
					},
				},
			},
		},
	}
	err := BOS_CLIENT.PutBucketNotificationFromStruct(EXISTS_BUCKET, args)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketNotification(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, len(res.Notifications), 1)
	notification := res.Notifications[0]
	ExpectEqual(t.Errorf, notification.Id, "id-1")
	ExpectEqual(t.Errorf, notification.AppId, "app-id-1")
	ExpectEqual(t.Errorf, notification.Status, "enabled")
	ExpectEqual(t.Errorf, notification.Resources, []string{"/path1/*", "/*.jpg", "/*.png"})
	ExpectEqual(t.Errorf, notification.Events, []string{"PutObject"})
	ExpectEqual(t.Errorf, notification.Quota.QuotaDay, float64(1000))
	ExpectEqual(t.Errorf, notification.Quota.QuotaSec, float64(10))
	ExpectEqual(t.Errorf, notification.Apps[0].Id, "app-id-1")
	ExpectEqual(t.Errorf, notification.Apps[0].EventUrl, "http://127.0.0.1:8080/event")

	err = BOS_CLIENT.PutBucketNotificationFromStruct(EXISTS_BUCKET, nil)
	ExpectEqual(t.Errorf, err != nil, true)
}

func TestDeleteNotification(t *testing.T) {
	err := BOS_CLIENT.DeleteBucketNotification(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetBucketNotification(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err != nil, true)
	ExpectEqual(t.Errorf, res == nil, true)
}

func TestPutObject(t *testing.T) {
	args := &api.PutObjectArgs{StorageClass: api.STORAGE_CLASS_COLD}
	body, _ := bce.NewBodyFromString("aaaaaaaaaaa")
	res, err := BOS_CLIENT.PutObject(EXISTS_BUCKET, "test-put-object", body, args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("etag: %v", res)
}

func TestBasicPutObject(t *testing.T) {
	body, _ := bce.NewBodyFromString("aaaaaaaaaaa")
	res, err := BOS_CLIENT.BasicPutObject(EXISTS_BUCKET, "test-put-object", body)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("etag: %v", res)
}

func TestPutObjectFromBytes(t *testing.T) {
	arr := []byte("123")
	res, err := BOS_CLIENT.PutObjectFromBytes(EXISTS_BUCKET, "test-put-object", arr, nil)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("etag: %v", res)
}

func TestPutObjectFromString(t *testing.T) {
	res, err := BOS_CLIENT.PutObjectFromString(EXISTS_BUCKET, "test-put-object", "123", nil)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("etag: %v", res)
}

func TestPutObjectFromFile(t *testing.T) {
	fname := "/tmp/test-put-file"
	f, _ := os.Create(fname)
	f.WriteString("12345")
	f.Close()
	res, err := BOS_CLIENT.PutObjectFromFile(EXISTS_BUCKET, "test-put-object", fname, nil)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("etag: %v", res)

	args := &api.PutObjectArgs{ContentLength: 6}
	res, err = BOS_CLIENT.PutObjectFromFile(EXISTS_BUCKET, "test-put-object", fname, args)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("error: %v", err)

	args.ContentLength = -1
	res, err = BOS_CLIENT.PutObjectFromFile(EXISTS_BUCKET, "test-put-object", fname, args)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("error: %v", err)
	os.Remove(fname)
}

func TestCopyObject(t *testing.T) {
	args := new(api.CopyObjectArgs)
	args.StorageClass = api.STORAGE_CLASS_COLD
	res, err := BOS_CLIENT.CopyObject(EXISTS_BUCKET, "test-copy-object",
		EXISTS_BUCKET, "test-put-object", args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("copy result: %+v", res)
}

func TestBasicCopyObject(t *testing.T) {
	res, err := BOS_CLIENT.BasicCopyObject(EXISTS_BUCKET, "test-copy-object",
		EXISTS_BUCKET, "test-put-object")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("copy result: %+v", res)
}

func TestGetObject(t *testing.T) {
	res, err := BOS_CLIENT.GetObject(EXISTS_BUCKET, "test-put-object", nil)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
	t.Logf("%v", res.ContentLength)
	buf := make([]byte, 1024)
	n, _ := res.Body.Read(buf)
	t.Logf("%s", buf[0:n])
	res.Body.Close()

	respHeaders := map[string]string{"ContentEncoding": "qqqqqqqqqqqqq"}
	res, err = BOS_CLIENT.GetObject(EXISTS_BUCKET, "test-put-object", respHeaders)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
	t.Logf("%v", res.ContentLength)
	n, _ = res.Body.Read(buf)
	t.Logf("%s", buf[0:n])
	res.Body.Close()

	res, err = BOS_CLIENT.GetObject(EXISTS_BUCKET, "test-put-object", respHeaders, 2)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
	t.Logf("%v", res.ContentLength)
	n, _ = res.Body.Read(buf)
	t.Logf("%s", buf[0:n])

	res, err = BOS_CLIENT.GetObject(EXISTS_BUCKET, "test-put-object", respHeaders, 2, 4)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
	t.Logf("%v", res.ContentLength)
	n, _ = res.Body.Read(buf)
	t.Logf("%s", buf[0:n])
	res.Body.Close()
}

func TestBasicGetObject(t *testing.T) {
	res, err := BOS_CLIENT.BasicGetObject(EXISTS_BUCKET, "test-put-object")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)

	defer res.Body.Close()
	t.Logf("%v", res.ContentLength)
	for {
		buf := make([]byte, 1024)
		n, e := res.Body.Read(buf)
		t.Logf("%s", buf[0:n])
		if e != nil {
			break
		}
	}
}

func TestBasicGetObjectToFile(t *testing.T) {
	fname := "/tmp/test-get-object"
	err := BOS_CLIENT.BasicGetObjectToFile(EXISTS_BUCKET, "test-put-object", fname)
	ExpectEqual(t.Errorf, err, nil)
	os.Remove(fname)

	fname = "/bin/test-get-object"
	err = BOS_CLIENT.BasicGetObjectToFile(EXISTS_BUCKET, "test-put-object", fname)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v", err)

	err = BOS_CLIENT.BasicGetObjectToFile(EXISTS_BUCKET, "not-exist-object-name", fname)
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v", err)
}

func TestGetObjectMeta(t *testing.T) {
	res, err := BOS_CLIENT.GetObjectMeta(EXISTS_BUCKET, "test-put-object")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("get object meta result: %+v", res)
}

func TestFetchObject(t *testing.T) {
	args := &api.FetchObjectArgs{api.FETCH_MODE_ASYNC, api.STORAGE_CLASS_COLD}
	res, err := BOS_CLIENT.FetchObject(EXISTS_BUCKET, "test-fetch-object",
		"https://cloud.baidu.com/doc/BOS/API.html", args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("result: %+v", res)
}

func TestBasicFetchObject(t *testing.T) {
	res, err := BOS_CLIENT.BasicFetchObject(EXISTS_BUCKET, "test-fetch-object",
		"https://cloud.baidu.com/doc/BOS/API.html")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("result: %+v", res)

	res1, err1 := BOS_CLIENT.GetObjectMeta(EXISTS_BUCKET, "test-fetch-object")
	ExpectEqual(t.Errorf, err1, nil)
	t.Logf("meta: %+v", res1)
}

func TestSimpleFetchObject(t *testing.T) {
	res, err := BOS_CLIENT.SimpleFetchObject(EXISTS_BUCKET, "test-fetch-object",
		"https://cloud.baidu.com/doc/BOS/API.html",
		api.FETCH_MODE_ASYNC, api.STORAGE_CLASS_COLD)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("result: %+v", res)
}

func TestAppendObject(t *testing.T) {
	args := &api.AppendObjectArgs{}
	body, _ := bce.NewBodyFromString("aaaaaaaaaaa")
	res, err := BOS_CLIENT.AppendObject(EXISTS_BUCKET, "test-append-object", body, args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestSimpleAppendObject(t *testing.T) {
	body, _ := bce.NewBodyFromString("bbbbbbbbbbb")
	res, err := BOS_CLIENT.SimpleAppendObject(EXISTS_BUCKET, "test-append-object", body, 11)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestSimpleAppendObjectFromString(t *testing.T) {
	res, err := BOS_CLIENT.SimpleAppendObjectFromString(
		EXISTS_BUCKET, "test-append-object", "123", 22)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestSimpleAppendObjectFromFile(t *testing.T) {
	fname := "/tmp/test-append-file"
	f, _ := os.Create(fname)
	f.WriteString("12345")
	f.Close()
	res, err := BOS_CLIENT.SimpleAppendObjectFromFile(EXISTS_BUCKET, "test-append-object", fname, 25)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
	os.Remove(fname)
}

func TestDeleteObject(t *testing.T) {
	err := BOS_CLIENT.DeleteObject(EXISTS_BUCKET, "test-put-object")
	ExpectEqual(t.Errorf, err, nil)
}

func TestDeleteMultipleObjectsFromString(t *testing.T) {
	multiDeleteStr := `{
    "objects":[
        {"key": "aaaa"},
        {"key": "test-copy-object"},
        {"key": "test-append-object"},
        {"key": "cccc"}
    ]
}`
	res, err := BOS_CLIENT.DeleteMultipleObjectsFromString(EXISTS_BUCKET, multiDeleteStr)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestDeleteMultipleObjectsFromStruct(t *testing.T) {
	multiDeleteObj := &api.DeleteMultipleObjectsArgs{[]api.DeleteObjectArgs{
		api.DeleteObjectArgs{"1"}, api.DeleteObjectArgs{"test-fetch-object"}}}
	res, err := BOS_CLIENT.DeleteMultipleObjectsFromStruct(EXISTS_BUCKET, multiDeleteObj)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestDeleteMultipleObjectsFromKeyList(t *testing.T) {
	keyList := []string{"aaaa", "test-copy-object", "test-append-object", "cccc"}
	res, err := BOS_CLIENT.DeleteMultipleObjectsFromKeyList(EXISTS_BUCKET, keyList)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestInitiateMultipartUpload(t *testing.T) {
	args := &api.InitiateMultipartUploadArgs{Expires: "aaaaaaa"}
	res, err := BOS_CLIENT.InitiateMultipartUpload(EXISTS_BUCKET, "test-multipart-upload", "", args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)

	err1 := BOS_CLIENT.AbortMultipartUpload(EXISTS_BUCKET,
		"test-multipart-upload", res.UploadId)
	ExpectEqual(t.Errorf, err1, nil)
}

func TestBasicInitiateMultipartUpload(t *testing.T) {
	res, err := BOS_CLIENT.BasicInitiateMultipartUpload(EXISTS_BUCKET, "test-multipart-upload")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)

	err1 := BOS_CLIENT.AbortMultipartUpload(EXISTS_BUCKET,
		"test-multipart-upload", res.UploadId)
	ExpectEqual(t.Errorf, err1, nil)
}

func TestUploadPart(t *testing.T) {
	res, err := BOS_CLIENT.UploadPart(EXISTS_BUCKET, "a", "b", 1, nil, nil)
	t.Logf("%+v, %+v", res, err)
}

func TestUploadPartCopy(t *testing.T) {
	res, err := BOS_CLIENT.UploadPartCopy(EXISTS_BUCKET, "test-multipart-upload",
		EXISTS_BUCKET, "test-multipart-copy", "12345", 1, nil)
	t.Logf("%+v, %+v", res, err)
}

func TestBasicUploadPartCopy(t *testing.T) {
	res, err := BOS_CLIENT.BasicUploadPartCopy(EXISTS_BUCKET, "test-multipart-upload",
		EXISTS_BUCKET, "test-multipart-copy", "12345", 1)
	t.Logf("%+v, %+v", res, err)
}

func TestListMultipartUploads(t *testing.T) {
	args := &api.ListMultipartUploadsArgs{MaxUploads: 10}
	res, err := BOS_CLIENT.ListMultipartUploads(EXISTS_BUCKET, args)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestBasicListMultipartUploads(t *testing.T) {
	res, err := BOS_CLIENT.BasicListMultipartUploads(EXISTS_BUCKET)
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%+v", res)
}

func TestUploadSuperFile(t *testing.T) {
	err := BOS_CLIENT.UploadSuperFile(EXISTS_BUCKET, "super-object", "/tmp/super-file", "")
	ExpectEqual(t.Errorf, err, nil)

	err = BOS_CLIENT.UploadSuperFile(EXISTS_BUCKET, "not-exist", "not-exist", "")
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%+v", err)
}

func TestDownloadSuperFile(t *testing.T) {
	err := BOS_CLIENT.DownloadSuperFile(EXISTS_BUCKET, "super-object", "/tmp/download-super-file")
	ExpectEqual(t.Errorf, err, nil)

	err = BOS_CLIENT.DownloadSuperFile(EXISTS_BUCKET, "not-exist", "/tmp/not-exist")
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%+v", err)
}

func TestGeneratePresignedUrl(t *testing.T) {
	url := BOS_CLIENT.BasicGeneratePresignedUrl(EXISTS_BUCKET, "glog.go", 100)
	resp, err := http.Get(url)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, resp.StatusCode, 200)

	params := map[string]string{"responseContentType": "text"}
	url = BOS_CLIENT.GeneratePresignedUrl(EXISTS_BUCKET, "glog.go", 100, "HEAD", nil, params)
	resp, err = http.Head(url)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, resp.StatusCode, 200)

	BOS_CLIENT.Config.Endpoint = "10.180.112.31"
	url = BOS_CLIENT.GeneratePresignedUrl(EXISTS_BUCKET, "glog.go", 100, "HEAD", nil, params)
	resp, err = http.Head(url)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, resp.StatusCode, 200)

	BOS_CLIENT.Config.Endpoint = "10.180.112.31:80"
	url = BOS_CLIENT.GeneratePresignedUrl(EXISTS_BUCKET, "glog.go", 100, "HEAD", nil, params)
	resp, err = http.Head(url)
	ExpectEqual(t.Errorf, err, nil)
	ExpectEqual(t.Errorf, resp.StatusCode, 200)
}

func TestPutObjectAcl(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[{
                "id":"e13b12d0131b4c8bae959df4969387b8"
            }],
            "permission":["READ"]
        }
    ]
}`
	body, _ := bce.NewBodyFromString(acl)
	err := BOS_CLIENT.PutObjectAcl(EXISTS_BUCKET, "glog.go", body)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "READ")
}

func TestPutObjectAclFromCanned(t *testing.T) {
	err := BOS_CLIENT.PutObjectAclFromCanned(EXISTS_BUCKET, "glog.go", api.CANNED_ACL_PUBLIC_READ)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
}

func TestPutObjectAclGrantRead(t *testing.T) {
	err := BOS_CLIENT.PutObjectAclGrantRead(EXISTS_BUCKET,
		"glog.go", "e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "READ")
}

func TestPutObjectAclGrantFullControl(t *testing.T) {
	err := BOS_CLIENT.PutObjectAclGrantFullControl(EXISTS_BUCKET,
		"glog.go", "e13b12d0131b4c8bae959df4969387b8")
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutObjectAclFromFile(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[{
                "id":"e13b12d0131b4c8bae959df4969387b8"
            }],
            "permission":["READ"]
        }
    ]
}`
	fname := "/tmp/test-put-object-acl-by-file"
	f, _ := os.Create(fname)
	f.WriteString(acl)
	f.Close()
	err := BOS_CLIENT.PutObjectAclFromFile(EXISTS_BUCKET, "glog.go", fname)
	os.Remove(fname)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "READ")
}

func TestPutObjectAclFromString(t *testing.T) {
	acl := `{
    "accessControlList":[
        {
            "grantee":[{
                "id":"e13b12d0131b4c8bae959df4969387b8"
            }],
            "permission":["FULL_CONTROL"]
        }
    ]
}`
	err := BOS_CLIENT.PutObjectAclFromString(EXISTS_BUCKET, "glog.go", acl)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "FULL_CONTROL")
}

func TestPutObjectAclFromStruct(t *testing.T) {
	aclObj := &api.PutObjectAclArgs{
		[]api.GrantType{
			api.GrantType{
				Grantee: []api.GranteeType{
					api.GranteeType{"e13b12d0131b4c8bae959df4969387b8"},
				},
				Permission: []string{
					"READ",
				},
			},
		},
	}
	err := BOS_CLIENT.PutObjectAclFromStruct(EXISTS_BUCKET, "glog.go", aclObj)
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	t.Logf("%v", res)
	ExpectEqual(t.Errorf, res.AccessControlList[0].Permission[0], "READ")
}

func TestDeleteObjectAcl(t *testing.T) {
	err := BOS_CLIENT.DeleteObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err, nil)
	res, err := BOS_CLIENT.GetObjectAcl(EXISTS_BUCKET, "glog.go")
	ExpectEqual(t.Errorf, err != nil, true)
	t.Logf("%v, %v", res, err)
}