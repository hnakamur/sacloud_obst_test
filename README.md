# sacloud_objt_test

[オブジェクトストレージ AWS CLI および AWS SDK ご利用できないバージョンのご案内 | さくらのクラウドニュース](https://cloud.sakura.ad.jp/news/2025/02/04/objectstorage_defectversion/)
を読んで、[aws/aws-sdk-go-v2: AWS SDK for the Go programming language.](https://github.com/aws/aws-sdk-go-v2)での対応方法を調べた作業レポジトリです。

* v1.72.3以前は問題なし。
* v1.73.0～v1.74.0まではnil参照で落ちた。
  * 原因は調査していない。
* v1.74.1以降は意図せぬチェックサムがついてオブジェクトの内容が壊れる。
  * `github.com/aws/aws-sdk-go-v2/service/s3`の`Client`作成時に以下のようにオプションを指定すれば回避できた。

```
c, err := config.LoadDefaultConfig(ctx,
	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		cfg.AccessKey, cfg.Secret, "",
	)),
	config.WithRegion(cfg.Region),
)
if err != nil {
	return nil, err
}

svc := s3.NewFromConfig(c, func(o *s3.Options) {
	o.BaseEndpoint = aws.String(cfg.EndpointURL)
	o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
})
```

なお、今回実際に試したのは、v1.72.3、v1.73.0、v1.73.1、v1.73.2、v1.74.0、v1.74.1、v1.77.1のみです。
