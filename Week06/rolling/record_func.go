package rolling

func IncreaseSuccess(b *Bucket) {
	b.AddSuccess(1)
}

func IncreaseFailure(b *Bucket) {
	b.AddSuccess(1)
}
