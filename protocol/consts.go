package protocol

const (
	OP_REPLY        int32 = 1
	OP_UPDATE       int32 = 2001
	OP_INSERT       int32 = 2002
	RESERVED        int32 = 2003
	OP_QUERY        int32 = 2004
	OP_GET_MORE     int32 = 2005
	OP_DELETE       int32 = 2006
	OP_KILL_CURSORS int32 = 2007
	OP_COMMAND      int32 = 2010
	OP_MSG          int32 = 2013
)
