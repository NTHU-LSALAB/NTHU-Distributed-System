package daomock

//go:generate mockgen -destination=mock.go -package=$GOPACKAGE github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao VideoDAO
