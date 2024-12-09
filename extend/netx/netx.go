/**
 * @Author: dingQingHui
 * @Description:
 * @File: netx
 * @Version: 1.0.0
 * @Date: 2024/11/29 15:04
 */

package netx

import (
	"github.com/panjf2000/gnet/v2/pkg/errors"
	"strings"
)

func ParseProtoAddr(protoAddr string) (string, string, error) {
	protoAddr = strings.ToLower(protoAddr)
	if strings.Count(protoAddr, "://") != 1 {
		return "", "", errors.ErrInvalidNetworkAddress
	}
	pair := strings.SplitN(protoAddr, "://", 2)
	if len(pair) < 2 {
		return "", "", errors.ErrInvalidNetworkAddress
	}
	proto, addr := pair[0], pair[1]
	switch proto {
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "unix":
	default:
		return "", "", errors.ErrUnsupportedProtocol
	}
	if addr == "" {
		return "", "", errors.ErrInvalidNetworkAddress
	}
	return proto, addr, nil
}
