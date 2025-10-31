package main

/*
#cgo CFLAGS: -I/Users/cmll/Library/Java/JavaVirtualMachines/corretto-1.8.0_462-1/Contents/Home/include
#cgo CFLAGS: -I/Users/cmll/Library/Java/JavaVirtualMachines/corretto-1.8.0_462-1/Contents/Home/include/darwin
#cgo CFLAGS: -I../Layer4forwarding/layer4-common/target/generated-jni-headers
#cgo darwin LDFLAGS: -framework Security -framework CoreFoundation
#include <stdio.h>
#include <stdlib.h>
#include <com_net_layer4_common_netty_channel_Hysteria2ProxyChannel.h>

static const char* GetJStringUTF(JNIEnv* env, jstring jstr) {
    if (jstr == NULL) return NULL;
    return (*env)->GetStringUTFChars(env, jstr, NULL);
}

static void ReleaseJStringUTF(JNIEnv* env, jstring jstr, const char* cstr) {
    if (cstr != NULL) {
        (*env)->ReleaseStringUTFChars(env, jstr, cstr);
    }
}

static jobject NewGlobalRef(JNIEnv* env, jobject obj) {
    jobject gobj = (*env)->NewGlobalRef(env, obj);
    return gobj;
}

static void DeleteGlobalRef(JNIEnv* env, jobject gobj) {
    (*env)->DeleteGlobalRef(env, gobj);
}

static JNIEnv* GetJNIEnv(JavaVM* jvm) {
    JNIEnv *env = NULL;
    if ((*jvm)->GetEnv(jvm, (void**)&env, JNI_VERSION_1_6) != JNI_OK) {
        (*jvm)->AttachCurrentThread(jvm, (void**)&env, NULL);
    }
    return env;
}

static jmethodID GetMethodID(JNIEnv* env, jobject obj, const char *name, const char *sig) {
    jclass cls = (*env)->GetObjectClass(env, obj);
    jmethodID mid = (*env)->GetMethodID(env, cls, name, sig);
    return mid;
}

static jstring GetJMsg(JNIEnv* env, const char* errmsg) {
    jstring jmsg = NULL;
    if (errmsg != NULL) {
        jmsg = (*env)->NewStringUTF(env, errmsg);
    }
    return jmsg;
}

static void ReleaseObjs(JNIEnv* env, jobject obj, jstring jmsg) {
    if (jmsg != NULL) {
        (*env)->DeleteLocalRef(env, jmsg);
    }
    DeleteGlobalRef(env, obj);
}

static void callConnectResp(JavaVM* jvm, jobject obj, const char* errmsg, const long connectionId) {
    fprintf(stderr,"callConnectResp debug1\n");
    JNIEnv* env = GetJNIEnv(jvm);
    fprintf(stderr,"callConnectResp debug2\n");
    jstring jmsg = GetJMsg(env, errmsg);
    fprintf(stderr,"callConnectResp debug3\n");
    jmethodID mid = GetMethodID(env, obj, "connectResp", "(Ljava/lang/String;J)V");
    fprintf(stderr,"callConnectResp debug4\n");
    (*env)->CallVoidMethod(env, obj, mid, jmsg, (jlong)connectionId);
    fprintf(stderr,"callConnectResp debug5\n");
    ReleaseObjs(env, obj, jmsg);
}

*/
import "C"
import (
	"fmt"
	"net"
	"strconv"
	"time"
	"unsafe"

	"github.com/apernet/hysteria/app/v2/cmd"
	"github.com/apernet/hysteria/core/v2/client"
)

var gJvm *C.JavaVM
var hyPool *HYPool

//export JNI_OnLoad
func JNI_OnLoad(vm *C.JavaVM, reserved unsafe.Pointer) C.jint {
	fmt.Println("JNI_OnLoad")
	gJvm = vm
	hyPool = &HYPool{
		clients: make(map[string]*RefClient),
	}
	return C.JNI_VERSION_1_6
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq(env *C.JNIEnv, obj C.jobject,
	dhost C.jstring, dport C.jint,
	server C.jstring, password C.jstring, port C.jint, skipcertverify C.jboolean, sni C.jstring, udp C.jboolean) {
	gdhost := JString2Go(env, dhost)
	gserver := JString2Go(env, server)
	gpassword := JString2Go(env, password)
	gsni := JString2Go(env, sni)
	fmt.Println("connectReq begin", gdhost, dport, gserver, gpassword, port, skipcertverify, gsni, udp)
	gobj := C.NewGlobalRef(env, obj)
	go func() {

		hyaddr := net.JoinHostPort(gserver, strconv.Itoa(int(port)))
		fmt.Println("connectResp begin", gdhost, dport)

		hyc, err := hyPool.GetClient(hyaddr, func() (*client.Config, error) {
			hyConfig := &client.Config{}
			hostPort := net.JoinHostPort(gserver, strconv.Itoa(int(port)))
			addr, err := net.ResolveUDPAddr("udp", hostPort)
			if err != nil {
				return nil, err
			}
			hyConfig.ServerAddr = addr
			if gsni == "" {
				hyConfig.TLSConfig.ServerName = gserver
			} else {
				hyConfig.TLSConfig.ServerName = gsni
			}
			hyConfig.ConnFactory = &udpConnFactory{}
			hyConfig.Auth = gpassword
			hyConfig.TLSConfig.InsecureSkipVerify = JBoolean2Go(skipcertverify)
			return hyConfig, nil
		})

		if err != nil {
			C.callConnectResp(gJvm, gobj, GString2C(err.Error()), 0)
			return
		}
		defer func(hyPool *HYPool, addr string) {
			err := hyPool.ReleaseClient(addr)
			if err != nil {
			}
		}(hyPool, hyaddr)

		dstaddr := net.JoinHostPort(gdhost, strconv.Itoa(int(dport)))

		rConn, err := hyc.TCP(dstaddr)
		if err != nil {
			C.callConnectResp(gJvm, gobj, GString2C(err.Error()), 0)
			return
		}

		defer rConn.Close()

		time.Sleep(3 * time.Second)
		fmt.Println("connecting~~~~", gdhost, dport)
		C.callConnectResp(gJvm, gobj, nil, 111)
		fmt.Println("connectResp end", gdhost, dport)
	}()
	fmt.Println("connectReq end", gdhost, dport)
}

func JString2Go(env *C.JNIEnv, jstr C.jstring) string {
	cstr := C.GetJStringUTF(env, jstr)
	if cstr == nil {
		return ""
	}
	defer C.ReleaseJStringUTF(env, jstr, cstr)
	gstr := C.GoString(cstr)
	return gstr
}

func JBoolean2Go(b C.jboolean) bool {
	return b != 0
}

func GString2C(gstr string) *C.char {
	cerr := C.CString(gstr)
	defer C.free(unsafe.Pointer(cerr))
	return cerr
}

func connectLog(info *client.HandshakeInfo, count int) {
	fmt.Println("connected to server:", "udpEnabled=", info.UDPEnabled, ",tx=", info.Tx, ",count=", count)
}

type udpConnFactory struct{}

func (f *udpConnFactory) New(addr net.Addr) (net.PacketConn, error) {
	return net.ListenUDP("udp", nil)
}

func main() {
	//C.callConnectResp(gJvm, C.jobject(C.NULL), nil, 111) //
	cmd.Execute()
}
