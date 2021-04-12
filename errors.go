package client

import "errors"

var prefix = "len of field "
var suffixMore0 = " must be more 0"
var suffixEq0 = " must be equal 0"

var CountArgsError = errors.New("field countArgs must be equal length of the len(args)")
var PathError = errors.New("path not passed")
var EmptyMethodError = errors.New(prefix + "method" + suffixMore0)
var EmptyPathError = errors.New(prefix + "url" + suffixMore0)
var NotEmptyBodyError = errors.New(prefix + "body" + suffixEq0)
var CloseBodyError = errors.New("close method of body of response has error")
var ReadBodyError = errors.New("read method of body of response has error")
var UnmarshalResponseError = errors.New("unmarshalling response has error")
var UnmarshalRequestError = errors.New("unmarshalling request has error")
