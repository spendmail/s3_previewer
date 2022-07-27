package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	internalresizer "github.com/spendmail/s3_previewer/internal/resizer"
	"io/ioutil"
	"log"
)

func main() {

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	//get
	response, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String("nngfpgan"),
		Key:    aws.String("experiments/style_52_disney_3d_barbieken_y_def_0.75_64_1_50K_tr_psi_1.0_style_52_disney_3d_barbieken_y_def_0.75_64_1_50K_tr_psi_1.0_crops_fg_and_skin/visualization/005000/imgs/0aac46be717bf5ded72225677d6b01e84c87cea3.jpeg"),
	})
	// Make sure to always close the response Body when finished
	defer response.Body.Close()

	if err != nil {
		panic(err)
	}

	fmt.Println(*response.ContentType)
	body, err := ioutil.ReadAll(response.Body)

	resizer := internalresizer.New()
	resultBytes, err := resizer.Resize(1024, 0, body)
	if err != nil {
		panic(err)
	}

	//err = ioutil.WriteFile("/tmp/1.jpeg", body, 0644)
	err = ioutil.WriteFile("/tmp/3.jpeg", resultBytes, 0644)
	if err != nil {
		panic(err)
	}
}
