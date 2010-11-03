package core

import "msgId/msgId"

type ServiceMsg interface {
    Id() msgId.MsgId
}
