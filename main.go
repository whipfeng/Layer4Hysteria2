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
	"time"
	"unsafe"

	"github.com/apernet/hysteria/app/v2/cmd"
	"github.com/apernet/hysteria/core/v2/client"
	"go.uber.org/zap"
)

var gJvm *C.JavaVM
var logger *zap.Logger

//export JNI_OnLoad
func JNI_OnLoad(vm *C.JavaVM, reserved unsafe.Pointer) C.jint {
	fmt.Println("JNI_OnLoad")
	gJvm = vm
	return C.JNI_VERSION_1_6
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq(env *C.JNIEnv, obj C.jobject,
	dhost C.jstring, dport C.jint,
	server C.jstring, password C.jstring, port C.jint, skipcertverify C.jboolean, sni C.jstring, udp C.jboolean) {
	gdhost := java2gostr(env, dhost)
	gserver := java2gostr(env, server)
	gpassword := java2gostr(env, password)
	gsni := java2gostr(env, sni)
	fmt.Println("connectReq begin", gdhost, dport, gserver, gpassword, port, skipcertverify, gsni, udp)
	gobj := C.NewGlobalRef(env, obj)
	go func() {
		fmt.Println("connectResp begin", gdhost, dport)

		hyclient, err := client.NewReconnectableClient(
			func() (*client.Config, error) {
				hyConfig := &client.Config{}
				return hyConfig, nil
			},
			func(c client.Client, info *client.HandshakeInfo, count int) {
				connectLog(info, count)
			}, false)
		if err != nil {
			logger.Fatal("failed to initialize client", zap.Error(err))
		}
		defer hyclient.Close()

		time.Sleep(3 * time.Second)
		fmt.Println("connecting~~~~", gdhost, dport)
		C.callConnectResp(gJvm, gobj, nil, 111)
		fmt.Println("connectResp end", gdhost, dport)
	}()
	fmt.Println("connectReq end", gdhost, dport)
}

func java2gostr(env *C.JNIEnv, jstr C.jstring) string {
	cstr := C.GetJStringUTF(env, jstr)
	gstr := C.GoString(cstr)
	C.ReleaseJStringUTF(env, jstr, cstr)
	return gstr
}

func connectLog(info *client.HandshakeInfo, count int) {
	logger.Info("connected to server",
		zap.Bool("udpEnabled", info.UDPEnabled),
		zap.Uint64("tx", info.Tx),
		zap.Int("count", count))
}

func main() {
	//C.callConnectResp(gJvm, C.jobject(C.NULL), nil, 111) //
	cmd.Execute()
}
