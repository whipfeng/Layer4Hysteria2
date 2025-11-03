package main

/*
#cgo CFLAGS: -I/Users/cmll/Library/Java/JavaVirtualMachines/corretto-1.8.0_462-1/Contents/Home/include
#cgo CFLAGS: -I/Users/cmll/Library/Java/JavaVirtualMachines/corretto-1.8.0_462-1/Contents/Home/include/darwin
#cgo CFLAGS: -IC:/Users/AI/.jdks/corretto-1.8.0_462/include
#cgo CFLAGS: -IC:/Users/AI/.jdks/corretto-1.8.0_462/include/win32
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

static jfieldID GetfieldID(JNIEnv* env, jobject obj, const char *name, const char *sig) {
    jclass cls = (*env)->GetObjectClass(env, obj);
    jfieldID fid = (*env)->GetFieldID(env, cls, name, sig);
    return fid;
}

static jstring GetJMSGRef(JNIEnv* env, const char* msg) {
    jstring jmsg = NULL;
    if (msg != NULL) {
        jmsg = (*env)->NewStringUTF(env, msg);
    }
    return jmsg;
}

static void DelJMSGRef(JNIEnv* env, jstring jmsg) {
    if (jmsg != NULL) {
        (*env)->DeleteLocalRef(env, jmsg);
    }
}

static void SetConnectionID(JNIEnv* env, jobject obj, jlong connectionId) {
    jfieldID fid = GetfieldID(env, obj, "connectionId", "J");
    (*env)->SetLongField(env, obj, fid, connectionId);
}

static jlong GetConnectionID(JNIEnv* env, jobject obj) {
    jfieldID fid = GetfieldID(env, obj, "connectionId", "J");
    jlong connectionId = (*env)->GetLongField(env, obj, fid);
    return connectionId;
}

static void CallResp(JNIEnv* env, jobject obj, const char* name, jstring errmsg) {
    jmethodID mid = GetMethodID(env, obj, name, "(Ljava/lang/String;)V");
    (*env)->CallVoidMethod(env, obj, mid, errmsg);
}

static void CallReadResp(JNIEnv* env, jobject obj, jstring errmsg, jint len) {
    jmethodID mid = GetMethodID(env, obj, "readResp", "(Ljava/lang/String;I)V");
    (*env)->CallVoidMethod(env, obj, mid, errmsg, len);
}

*/
import "C"
import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"unsafe"

	"github.com/apernet/hysteria/core/v2/client"
)

var gjvm *C.JavaVM
var hyPool *HYPool

//export JNI_OnLoad
func JNI_OnLoad(vm *C.JavaVM, reserved unsafe.Pointer) C.jint {
	fmt.Println("JNI_OnLoad")
	gjvm = vm
	hyPool = NewHYPool()
	return C.JNI_VERSION_1_6
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_connectReq(env *C.JNIEnv, obj C.jobject,
	dhost C.jstring, dport C.jint,
	name C.jstring, server C.jstring, password C.jstring, port C.jint,
	skipcertverify C.jboolean, sni C.jstring, udp C.jboolean) {

	gname := J2GString(env, name)
	gdhost := J2GString(env, dhost)
	gserver := J2GString(env, server)
	gpassword := J2GString(env, password)
	gsni := J2GString(env, sni)
	fmt.Println("connectReq begin", gdhost, dport, gserver, gpassword, port, skipcertverify, gsni, udp)

	JNIEnvGoFunc(env, obj, func(genv *C.JNIEnv, gobj C.jobject) {

		hyaddr := net.JoinHostPort(gserver, strconv.Itoa(int(port)))
		dstaddr := net.JoinHostPort(gdhost, strconv.Itoa(int(dport)))
		fmt.Println("connecting begin", hyaddr, dstaddr)

		hd, err := hyPool.TCP(gname, func() (*client.Config, error) {
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
			hyConfig.TLSConfig.InsecureSkipVerify = J2GBoolean(skipcertverify)
			return hyConfig, nil
		}, dstaddr)

		if err == nil {
			C.SetConnectionID(genv, gobj, C.jlong(hd))
		}
		fmt.Println("connecting end", hyaddr, dstaddr, hd, err)
		fmt.Println("connectResp begin", hd)
		CallResp(genv, gobj, "connectResp", err)
		fmt.Println("connectResp end", hd)
	})
	fmt.Println("connectReq end", gdhost, dport, gserver, gpassword, port, skipcertverify, gsni, udp)
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_readReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_readReq(env *C.JNIEnv, obj C.jobject,
	addr C.jlong, len C.jint) {
	jhd := C.GetConnectionID(env, obj)
	hd := int64(jhd)
	glen := int(len)
	fmt.Println("readReq begin", hd)
	JNIEnvGoFunc(env, obj, func(genv *C.JNIEnv, gobj C.jobject) {
		fmt.Println("reading begin", hd)
		buf := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(addr))), glen)
		conn, err := hyPool.GetConn(hd)
		if err == nil {
			glen, err = conn.Read(buf)
		}
		fmt.Println("reading end", hd, glen, err)
		fmt.Println("readResp begin", hd)
		CallSelfResp(genv, gobj, err,
			func(env *C.JNIEnv, obj C.jobject, jemsg C.jstring) {
				C.CallReadResp(env, obj, jemsg, C.jint(glen))
			})
		fmt.Println("readResp end", hd)
	})
	fmt.Println("readReq end", hd)
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_writeReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_writeReq(env *C.JNIEnv, obj C.jobject,
	addr C.jlong, len C.jint) {
	jhd := C.GetConnectionID(env, obj)
	hd := int64(jhd)
	glen := int(len)
	fmt.Println("writeReq begin", hd, glen)
	JNIEnvGoFunc(env, obj, func(genv *C.JNIEnv, gobj C.jobject) {
		fmt.Println("writing begin", hd, glen)
		buf := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(addr))), glen)
		conn, err := hyPool.GetConn(hd)
		if err == nil {
			wlen, werr := conn.Write(buf)
			if werr == nil {
				if wlen != glen {
					err = fmt.Errorf("short write! %d", hd)
				}
			} else {
				err = werr
			}
		}
		fmt.Println("writing end", hd, glen, err)
		fmt.Println("writeResp begin", hd)
		CallResp(genv, gobj, "writeResp", err)
		fmt.Println("writeResp end", hd)
	})
	fmt.Println("writeReq end", hd)
}

//export Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_closeReq
func Java_com_net_layer4_common_netty_channel_Hysteria2ProxyChannel_closeReq(env *C.JNIEnv, obj C.jobject) {
	jhd := C.GetConnectionID(env, obj)
	hd := int64(jhd)
	fmt.Println("closeReq begin", hd)
	JNIEnvGoFunc(env, obj, func(genv *C.JNIEnv, gobj C.jobject) {
		fmt.Println("closing begin", hd)
		err := hyPool.Close(hd)
		fmt.Println("closing end", hd, err)
		fmt.Println("closeResp begin", hd)
		CallResp(genv, gobj, "closeResp", err)
		fmt.Println("closeResp end", hd)
	})
	fmt.Println("closeReq end", hd)
}

func JNIEnvGoFunc(env *C.JNIEnv, obj C.jobject, fn func(genv *C.JNIEnv, gobj C.jobject)) {
	gobj := C.NewGlobalRef(env, obj)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		genv := C.GetJNIEnv(gjvm)
		fn(genv, gobj)
		C.DeleteGlobalRef(genv, gobj)
	}()
}

func CallResp(env *C.JNIEnv, obj C.jobject, name string, emsg error) {
	CallSelfResp(env, obj, emsg,
		func(env *C.JNIEnv, obj C.jobject, jemsg C.jstring) {
			cname := C.CString(name)
			defer C.free(unsafe.Pointer(cname))
			C.CallResp(env, obj, cname, jemsg)
		})
}

func CallSelfResp(env *C.JNIEnv, obj C.jobject, emsg error,
	fn func(env *C.JNIEnv, obj C.jobject, jemsg C.jstring)) {

	if emsg == nil {
		fn(env, obj, C.GetJMSGRef(env, nil))
		return
	}
	cemsg := C.CString(emsg.Error())
	defer C.free(unsafe.Pointer(cemsg))
	jemsg := C.GetJMSGRef(env, cemsg)
	defer C.DelJMSGRef(env, jemsg)
	fn(env, obj, jemsg)
}

func G2JString(env *C.JNIEnv, gstr string, fn func(C.jstring)) {
	cstr := C.CString(gstr)
	defer C.free(unsafe.Pointer(cstr))
	jstr := C.GetJMSGRef(env, cstr)
	defer C.DelJMSGRef(env, jstr)
	fn(jstr)
}

func J2GString(env *C.JNIEnv, jstr C.jstring) string {
	cstr := C.GetJStringUTF(env, jstr)
	if cstr == nil {
		return ""
	}
	defer C.ReleaseJStringUTF(env, jstr, cstr)
	gstr := C.GoString(cstr)
	return gstr
}

func J2GBoolean(jb C.jboolean) bool {
	return jb != 0
}
