package core

import "msgId/msgId"

type ServiceMsg interface {
    Id() msgId.MsgId
}

type Service interface {
    Run(input chan ServiceMsg)
}
