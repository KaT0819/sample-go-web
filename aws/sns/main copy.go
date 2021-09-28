package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	AwsAccessKey       = os.Getenv("AwsAccessKey")
	AwsSecretAccessKey = os.Getenv("AwsSecretAccessKey")
	AwsRegion          = "us-east-1"
	AppId              = os.Getenv("AppId") // PinpointのプロジェクトID

	TopicARN = os.Getenv("TopicARN")
	SenderId = "testsend" // 送信ID SMSの送信者名

	PHONE = os.Getenv("PHONE")
	EMAIL = os.Getenv("EMAIL")

	// for Vonage
	VONAGE_API_KEY    = os.Getenv("VONAGE_API_KEY")
	VONAGE_API_SECRET = os.Getenv("VONAGE_API_SECRET")
)

func main() {

	// AWS SNSにアクセス

	sendSms()
	sendSns()
	sendPinpoint()
	sendVonage()
	sendgridmail()
}

func getClient() *sns.SNS {
	// mySession := session.Must(session.NewSession())
	// svc := sns.New(mySession, aws.NewConfig().WithRegion(AwsRegion))

	// クライアントの生成
	creds := credentials.NewStaticCredentials(AwsAccessKey, AwsSecretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(AwsRegion),
	})

	if err != nil {
		log.Fatal(err)
	}

	return sns.New(sess)
}

func getPinpointClient() *pinpoint.Pinpoint {
	// クライアントの生成
	creds := credentials.NewStaticCredentials(AwsAccessKey, AwsSecretAccessKey, "")
	sess, _ := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(AwsRegion),
	})

	return pinpoint.New(sess)
}

func sendSms() {
	log.Printf("sendsms start")
	// クライアントの生成
	client := getClient()

	// メッセージの作成
	pin := &sns.PublishInput{}
	pin.SetMessage("SMS エンドポイントのサンプルメッセージ")
	// 電話番号に国コードを指定します。今回は日本の場合は、[+81]を設定します。
	pin.SetPhoneNumber(PHONE)

	// SMS送信
	result, err := client.Publish(pin)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("result: %+v", result)
}

// SNS
func sendSns() {
	log.Printf("sendSnsMail start")

	// クライアントの生成
	client := getClient()

	// メッセージの作成
	text := "共通のサンプルメッセージです。"

	// サブスクリプションのプロトコルごとにメッセージを指定
	messageJson := map[string]string{
		"default": text,
		"sms":     "SMSへのサンプルメッセージ\n" + text,
		"email":   "Eメールへのサンプルメッセージ\n" + text,
	}
	// メッセージ構造体はJSON文字列にする
	bytes, err := json.Marshal(messageJson)
	if err != nil {
		fmt.Println("JSON marshal Error: ", err)
	}
	message := string(bytes)
	// log.Println(bytes)

	pin := &sns.PublishInput{
		Message:          aws.String(message),
		MessageStructure: aws.String("json"), // MessageStructureにjsonを指定
		TopicArn:         aws.String(TopicARN),
	}

	// 配信
	result, err := client.Publish(pin)
	if err != nil {
		fmt.Println("Publish Error: ", err)
	}

	// fmt.Println(result)
	log.Println(result.GoString())

}

// // Pinpoint
func sendPinpoint() {
	log.Printf("sendPinpoint start")

	// クライアント作成
	client := getPinpointClient()

	// メッセージ作成
	text := "Pinpoint サンプルメッセージ"

	// Pinpoint送信
	pin := &pinpoint.SendMessagesInput{
		ApplicationId: aws.String(AppId),
		MessageRequest: &pinpoint.MessageRequest{
			Addresses: map[string]*pinpoint.AddressConfiguration{
				PHONE: &pinpoint.AddressConfiguration{ // 電話番号を指定
					ChannelType: aws.String(pinpoint.ChannelTypeSms),
				},
				EMAIL: &pinpoint.AddressConfiguration{
					ChannelType: aws.String(pinpoint.ChannelTypeEmail),
				},
			},
			MessageConfiguration: &pinpoint.DirectMessageConfiguration{
				SMSMessage: &pinpoint.SMSMessage{
					Body:        aws.String(text),     // 本文
					SenderId:    aws.String(SenderId), // 送信ID SMSの送信者名
					MessageType: aws.String(pinpoint.MessageTypePromotional),
				},
				EmailMessage: &pinpoint.EmailMessage{
					FromAddress: aws.String(EMAIL), // メアドを設定
					SimpleEmail: &pinpoint.SimpleEmail{
						// 件名
						Subject: &pinpoint.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String("subject"),
						},
						// HTML本文
						HtmlPart: &pinpoint.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String("<HTML>html message</html>"),
						},
						// テキスト本文
						TextPart: &pinpoint.SimpleEmailPart{
							Charset: aws.String("utf-8"),
							Data:    aws.String("text message"),
						},
					},
				},
			},
		},
	}

	result, _ := client.SendMessages(pin)
	log.Printf("%+v", result)
}

// Vonage
func sendVonage() {
	// パラメータの作成
	value := url.Values{}
	value.Set("from", SenderId)
	value.Add("text", "サンプルメッセージ By Vonage API")
	value.Add("to", PHONE)
	value.Add("api_key", VONAGE_API_KEY)
	value.Add("api_secret", VONAGE_API_SECRET)
	value.Add("type", "unicode")

	// APIリクエスト
	resp, err := http.PostForm("https://rest.nexmo.com/sms/json", value)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 1024)

	respLen, _ := resp.Body.Read(buffer)
	body := string(buffer[:respLen])
	log.Println(body)
	log.Println(resp.Status)
	defer resp.Body.Close()
}

// Sendgrid
func sendgridmail() {
	// クライアント作成
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	// メッセージ作成
	from := mail.NewEmail("Example User", "test@example.com") // 面倒なのでFromはToと同じ
	subject := "サンプルのお知らせ Sendgrid"
	to := mail.NewEmail("Example User", EMAIL)
	plainTextContent := "サンプルテキストメッセージの送信"
	htmlContent := "<strong>サンプルテキストメッセージの送信</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)

	// メール送信
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}

// func customAuth() {
// 	log.Printf("customAuth start")
// 	// Admin SDK
// 	app, err := firebase.NewApp(context.Background(), nil)
// 	if err != nil {
// 		log.Fatalf("error initializing app: %v\n", err)
// 	}
// 	// opt := option.WithCredentialsFile("firebase-admin-account.json")
// 	// app, err := firebase.NewApp(context.Background(), nil, opt)
// 	// if err != nil {
// 	// 	log.Fatalf("error initializing app: %v", err)
// 	// }

// 	// type App struct {
// 	// 	authOverride     map[string]interface{}
// 	// 	dbURL            string
// 	// 	projectID        string
// 	// 	serviceAccountID string
// 	// 	storageBucket    string
// 	// 	opts             []option.ClientOption
// 	// }
// 	// Get an auth client from the firebase.App

// 	client, err := app.Auth(context.Background())
// 	if err != nil {
// 		log.Fatalf("error getting Auth client: %v\n", err)
// 	}

// 	email := "katuyuki.toki@gmail.com"
// 	getUserByEmail(client, email)
// 	// log.Printf("Successfully fetched user UID: %+v\n", u.UID)

// 	// u, err := client.GetUser(context.Background(), uid)
// 	// if err != nil {
// 	// 	log.Fatalf("error getting user %s: %v\n", uid, err)
// 	// }
// 	// log.Printf("Successfully fetched user data: %v\n", u)

// }

// // カスタムトークンの生成
// func getCustomToken() {
// 	log.Printf("getCustomToken start")

// 	// サービスアカウントのアドレスを指定してログインしている。
// 	// Permission iam.serviceAccounts.signBlob
// 	conf := &firebase.Config{
// 		ServiceAccountID: "firebase-adminsdk-kg307@new-3rdwatch-perftest.iam.gserviceaccount.com",
// 	}
// 	app, err := firebase.NewApp(context.Background(), conf)
// 	if err != nil {
// 		log.Fatalf("error getting Auth client: %v\n", err)
// 	}

// 	client, err := app.Auth(context.Background())
// 	if err != nil {
// 		log.Fatalf("error getting Auth client: %v\n", err)
// 	}

// 	uid := "Ws2FJQ0CBWhiJlux9nbyBd3WKVs2"

// 	token, err := client.CustomToken(context.Background(), uid)
// 	if err != nil {
// 		log.Fatalf("error minting custom token: %v\n", err)
// 	}

// 	log.Printf("Got custom token: %v\n", token)

// }

// func getUserByEmail(client *auth.Client, email string) *auth.UserRecord {
// 	u, err := client.GetUserByEmail(context.Background(), email)
// 	if err != nil {
// 		log.Fatalf("error getting user by email %s: %v\n", email, err)
// 	}
// 	log.Printf("Successfully user data: %#v\n", u)
// 	log.Printf(" UserInfo: %#v\n", u.UserInfo)
// 	log.Printf(" CustomClaims: %#v\n", u.CustomClaims)
// 	log.Printf(" ProviderUserInfo: %#v\n", u.ProviderUserInfo)

// 	return u
// }
