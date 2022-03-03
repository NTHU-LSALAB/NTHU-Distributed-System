package pbmock

//go:generate mockgen -destination=mock.go -package=$GOPACKAGE github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb CommentClient
