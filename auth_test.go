package czmq

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestAuthIPAllow(t *testing.T) {
	auth := NewAuth()
	defer auth.Destroy()

	var err error

	if testing.Verbose() {
		err = auth.Verbose()
		if err != nil {
			t.Error(err)
		}
	}

	err = auth.Allow("127.0.0.1")
	if err != nil {
		t.Error(err)
	}

	server := NewSock(Pull)
	server.SetOption(SockSetZapDomain("global"))
	defer server.Destroy()
	port, err := server.Bind("tcp://127.0.0.1:*")
	if err != nil {
		t.Error(err)
	}

	goodClient := NewSock(Push, SockSetZapDomain("global"))
	defer goodClient.Destroy()
	err = goodClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	badClient := NewSock(Push, SockSetZapDomain("global"))
	defer badClient.Destroy()
	err = badClient.Connect(fmt.Sprintf("tcp://127.0.0.2:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("World"), 0)
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("Bad World"), 0)
	if err != nil {
		t.Error(err)
	}

	poller, err := NewPoller(server)
	if err != nil {
		t.Error(err)
	}
	defer poller.Destroy()

	s, err := poller.Wait(200)
	if err != nil {
		t.Error(err)
	}
	if want, have := server, s; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	msg, err := s.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, have := "Hello", string(msg[0]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(msg[1]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	s, err = poller.Wait(200)
	if err != nil {
		t.Error(err)
	}
	if s != nil {
		t.Errorf("want %#v, have %#v", nil, s)
	}
}

func TestAuthPlain(t *testing.T) {
	auth := NewAuth()
	defer auth.Destroy()

	var err error

	if testing.Verbose() {
		err = auth.Verbose()
		if err != nil {
			t.Error(err)
		}
	}

	err = auth.Allow("127.0.0.1")
	if err != nil {
		t.Error(err)
	}

	file, err := os.Create("password.txt")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		os.Remove("password.txt")
	}()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString("admin=Password\n")
	if err != nil {
		t.Error(err)
	}
	writer.Flush()
	file.Close()

	err = auth.Plain("./password.txt")
	if err != nil {
		t.Error(err)
	}

	server := NewSock(Pull, SockSetZapDomain("global"))
	defer server.Destroy()
	server.SetOption(SockSetPlainServer(1))
	port, err := server.Bind("tcp://127.0.0.1:*")
	if err != nil {
		t.Error(err)
	}

	goodClient := NewSock(Push)
	defer goodClient.Destroy()
	goodClient.SetOption(SockSetPlainUsername("admin"))
	goodClient.SetOption(SockSetPlainPassword("Password"))
	goodClient.SetOption(SockSetZapDomain("global"))

	err = goodClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	badClient := NewSock(Push)
	defer badClient.Destroy()
	badClient.SetOption(SockSetPlainUsername("admin"))
	badClient.SetOption(SockSetPlainPassword("BadPassword"))
	badClient.SetOption(SockSetZapDomain("global"))

	err = badClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("World"), 0)
	if err != nil {
		t.Error(err)
	}

	poller, err := NewPoller(server)
	if err != nil {
		t.Error(err)
	}
	defer poller.Destroy()

	s, err := poller.Wait(2000)
	if err != nil {
		t.Error(err)
	}
	if want, have := server, s; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	msg, err := s.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, have := "Hello", string(msg[0]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(msg[1]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	err = badClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("World"), 0)
	if err != nil {
		t.Error(err)
	}

	s, err = poller.Wait(200)
	if err != nil {
		t.Error(err)
	}
	if s != nil {
		t.Errorf("want %#v, have %#v", nil, s)
	}

	if want, have := "Hello", string(msg[0]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(msg[1]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}
}

func TestAuthCurveAllowAny(t *testing.T) {
	// Skip CURVE tests on macOS in CI to avoid timeouts
	if runtime.GOOS == "darwin" && os.Getenv("CI") == "true" {
		t.Skip("Skipping CURVE test on macOS CI due to known timeout issues")
	}

	auth := NewAuth()
	defer auth.Destroy()

	var err error

	if testing.Verbose() {
		err = auth.Verbose()
		if err != nil {
			t.Error(err)
		}
	}

	server := NewSock(Pull, SockSetZapDomain("global"))
	defer server.Destroy()
	serverCert := NewCert()
	serverKey := serverCert.PublicText()
	serverCert.Apply(server)
	server.SetOption(SockSetCurveServer(1))

	goodClient := NewSock(Push)
	defer goodClient.Destroy()
	goodClientCert := NewCert()
	goodClientCert.Apply(goodClient)
	goodClient.SetOption(SockSetCurveServerkey(serverKey))

	badClient := NewSock(Push)
	defer badClient.Destroy()

	err = auth.Curve(CurveAllowAny)
	if err != nil {
		t.Error(err)
	}

	port, err := server.Bind("tcp://127.0.0.1:*")
	if err != nil {
		t.Error(err)
	}

	err = goodClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = badClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("World"), 0)
	if err != nil {
		t.Error(err)
	}

	poller, err := NewPoller(server)
	if err != nil {
		t.Error(err)
	}
	defer poller.Destroy()

	s, err := poller.Wait(2000)
	if err != nil {
		t.Error(err)
	}
	if want, have := server, s; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	msg, err := s.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, have := "Hello", string(msg[0]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(msg[1]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	err = badClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("Bad World"), 0)
	if err != nil {
		t.Error(err)
	}

	s, err = poller.Wait(200)
	if err != nil {
		t.Error(err)
	}
	if s != nil {
		t.Errorf("want %#v, have %#v", nil, s)
	}
}

func TestAuthCurveAllowCertificate(t *testing.T) {
	// Skip CURVE tests on macOS in CI to avoid timeouts
	if runtime.GOOS == "darwin" && os.Getenv("CI") == "true" {
		t.Skip("Skipping CURVE test on macOS CI due to known timeout issues")
	}

	testpath := path.Join("testauth")
	err := os.Mkdir(testpath, 0777)
	if err != nil {
		t.Error(err)
	}

	auth := NewAuth()
	defer auth.Destroy()

	if testing.Verbose() {
		err = auth.Verbose()
		if err != nil {
			t.Error(err)
		}
	}

	server := NewSock(Pull, SockSetZapDomain("global"))
	defer server.Destroy()
	serverCert := NewCert()
	serverKey := serverCert.PublicText()
	serverCert.Apply(server)
	server.SetOption(SockSetCurveServer(1))

	goodClient := NewSock(Push)
	defer goodClient.Destroy()
	goodClientCert := NewCert()
	goodClientCert.Apply(goodClient)
	goodClient.SetOption(SockSetCurveServerkey(serverKey))

	err = goodClientCert.SavePublic(path.Join(testpath, "goodClient"))
	if err != nil {
		t.Error(err)
	}

	badClient := NewSock(Push)
	defer badClient.Destroy()
	badClientCert := NewCert()
	badClientCert.Apply(badClient)
	badClient.SetOption(SockSetCurveServerkey(serverKey))

	err = auth.Curve(testpath)
	if err != nil {
		t.Error(err)
	}

	port, err := server.Bind("tcp://127.0.0.1:*")
	if err != nil {
		t.Error(err)
	}

	err = goodClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = badClient.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = goodClient.SendFrame([]byte("World"), 0)
	if err != nil {
		t.Error(err)
	}

	poller, err := NewPoller(server)
	if err != nil {
		t.Error(err)
	}
	defer poller.Destroy()

	s, err := poller.Wait(2000)
	if err != nil {
		t.Error(err)
	}
	if want, have := server, s; want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	msg, err := s.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, have := "Hello", string(msg[0]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	if want, have := "World", string(msg[1]); want != have {
		t.Errorf("want %#v, have %#v", want, have)
	}

	err = badClient.SendFrame([]byte("Hello"), 1)
	if err != nil {
		t.Error(err)
	}

	err = badClient.SendFrame([]byte("Bad World"), 0)
	if err != nil {
		t.Error(err)
	}

	s, err = poller.Wait(200)
	if err != nil {
		t.Error(err)
	}
	if s != nil {
		t.Errorf("want %#v, have %#v", nil, s)
	}

	os.RemoveAll(testpath)
}
