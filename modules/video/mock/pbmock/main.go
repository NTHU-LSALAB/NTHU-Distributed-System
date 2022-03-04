package pbmock

//go:generate mockgen -destination=mock.go -package=$GOPACKAGE github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb Video_UploadVideoServer,VideoClient
