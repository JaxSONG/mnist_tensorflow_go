package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	//	"strings"
	tf_core_framework "tensorflow/core/framework"
	pb "tensorflow_serving/apis"

	tg "github.com/galeone/tfgo"
	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
	//	"github.com/galeone/tfgo/image"
	"github.com/galeone/tfgo/image"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"

	"google.golang.org/grpc"
)

func main() {
	servingAddress := flag.String("serving-address", "localhost:9000", "The tensorflow serving address")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Usage: " + os.Args[0] + " --serving-address localhost:9000 path/to/img.png")
		os.Exit(1)
	}

	imgPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
	//imageBytes, err := ioutil.ReadFile(imgPath)
	_, err = ioutil.ReadFile(imgPath)

	if err != nil {
		log.Fatalln(err)
	}

	root := tg.NewRoot()
	img1 := image.ReadPNG(root, imgPath, 1)
	img1 = img1.ResizeArea(image.Size{Height: 28, Width: 28})
	results := tg.Exec(root, []tf.Output{img1.Value()}, nil, &tf.SessionOptions{})
	fmt.Println("len result:", len(results))
	tensor := results[0]

	fmt.Printf("DataType:%#v\n", tensor.DataType())
	fmt.Printf("shape:%#v\n", tensor.Shape())
	buf := new(bytes.Buffer)

	n, err := tensor.WriteContentsTo(buf)
	if err != nil {
		fmt.Println(err)
	}
	if n != int64(buf.Len()) {
		fmt.Printf(" WriteContentsTo said it wrote %v bytes, but wrote %v", n, buf.Len())
	}
	//    t2, err := ReadTensor(t1.DataType(), t1.Shape(), buf)
	t2, err := tf.ReadTensor(tensor.DataType(), []int64{28 * 28}, buf)
	if err != nil {
		fmt.Printf("%v", err)
	}
	fmt.Printf("DataType:%#v\n", t2.DataType())
	fmt.Printf("shape:%#v\n", t2.Shape())

	tensor = t2
	/*
		var tfoutput tf.Output
		tfoutput = img1.Value()
		fmt.Printf("shape:%#v", tfoutput.Shape)
		tfshape, err := tfoutput.Shape().ToSlice()
		if err != nil {
			panic(err)
		}
		tensor, err := tf.ReadTensor(tfoutput.DataType(), tfshape, strings.NewReader(string(imageBytes)))
	*/
	//	fmt.Println("value:", tensor.Value())
	request := &pb.PredictRequest{
		ModelSpec: &pb.ModelSpec{
			Name:          "mnist",
			SignatureName: "predict_images",
			Version: &google_protobuf.Int64Value{
				Value: int64(1),
			},
		},
		Inputs: map[string]*tf_core_framework.TensorProto{
			"images": &tf_core_framework.TensorProto{
				Dtype: tf_core_framework.DataType_DT_FLOAT,
				TensorShape: &tf_core_framework.TensorShapeProto{
					Dim: []*tf_core_framework.TensorShapeProto_Dim{
						&tf_core_framework.TensorShapeProto_Dim{
							Size: 1,
						},

						&tf_core_framework.TensorShapeProto_Dim{
							Size: 784,
						},
					},
				},
				//				FloatVal: tensor.Value().([]float32),
				FloatVal: tensor.Value().([]float32),
			},
		},
	}

	conn, err := grpc.Dial(*servingAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to the grpc server: %v\n", err)
	}
	defer conn.Close()

	client := pb.NewPredictionServiceClient(conn)

	resp, err := client.Predict(context.Background(), request)
	if err != nil {
		log.Fatalln("predict error:", err)
	}

	log.Println(resp)
}
func ToJsonString(v interface{}) string {
	re, _ := json.Marshal(v)
	return string(re)
}
