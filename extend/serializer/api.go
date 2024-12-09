/**
 * @Author: dingQingHui
 * @Description:
 * @File: api
 * @Version: 1.0.0
 * @Date: 2024/11/19 18:10
 */

package serializer

import "github.com/dingqinghui/gas/api"

var (
	Json    = api.ISerializer(new(jsonCodec))
	MsgPack = api.ISerializer(new(msgPackCodec))
	PB      = api.ISerializer(new(pbCodec))
)
