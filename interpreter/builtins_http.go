package interpreter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
)

func builtinHTTP(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "http expects request object"}
	}
	reqObj, ok := objectPairs(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "http expects object request"}
	}
	methodVal, ok := reqObj["method"]
	if !ok {
		methodVal = &String{Value: "GET"}
	}
	method, ok := stringArg(methodVal)
	if !ok {
		return nil, &RuntimeError{Message: "http method must be string"}
	}
	urlVal, ok := reqObj["url"]
	if !ok {
		return nil, &RuntimeError{Message: "http expects url"}
	}
	urlStr, ok := stringArg(urlVal)
	if !ok {
		return nil, &RuntimeError{Message: "http url must be string"}
	}
	var body io.Reader
	if bodyVal, ok := reqObj["body"]; ok && bodyVal != NullValue {
		bodyStr, ok := stringArg(bodyVal)
		if !ok {
			return nil, &RuntimeError{Message: "http body must be string"}
		}
		body = strings.NewReader(bodyStr)
	}

	reqDone := make(chan struct{})
	defer close(reqDone)

	ctx, cancel := context.WithCancel(context.Background())
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	fatalCh := runtimeFatalSignal(e)
	go func() {
		select {
		case <-cancelCh:
			cancel()
		case <-fatalCh:
			cancel()
		case <-reqDone:
			cancel()
		}
	}()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, recoverableError("http", "http request error: "+err.Error())
	}
	if headersVal, ok := reqObj["headers"]; ok && headersVal != NullValue {
		headers, err := extractHeaders(headersVal)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if cancelCh != nil {
				select {
				case <-cancelCh:
					return nil, canceledError()
				default:
				}
			}
			if fatalCh != nil {
				select {
				case <-fatalCh:
					return nil, runtimeFatalError(e)
				default:
				}
			}
			return nil, canceledError()
		}
		return nil, recoverableError("http", "http error: "+err.Error())
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, recoverableError("http", "http read error: "+err.Error())
	}
	headerMap := &Map{Pairs: make(map[MapKey]Value)}
	for key, values := range resp.Header {
		headerMap.Pairs[MapKey{Type: STRING, Value: key}] = &String{Value: strings.Join(values, ", ")}
	}
	return &Object{Pairs: map[string]Value{
		"status":  &Integer{Value: int64(resp.StatusCode)},
		"headers": headerMap,
		"body":    &String{Value: string(data)},
	}}, nil
}

func extractHeaders(val Value) (map[string]string, error) {
	switch headers := val.(type) {
	case *Object:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *ModuleObject:
		if headers.Env == nil {
			return nil, &RuntimeError{Message: "http headers must be object or map"}
		}
		out := make(map[string]string)
		for k, v := range headers.Env.Snapshot() {
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k] = str
		}
		return out, nil
	case *Map:
		out := make(map[string]string, len(headers.Pairs))
		for k, v := range headers.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "http headers keys must be strings"}
			}
			str, ok := stringArg(v)
			if !ok {
				return nil, &RuntimeError{Message: "http headers values must be strings"}
			}
			out[k.Value] = str
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "http headers must be object or map"}
	}
}
