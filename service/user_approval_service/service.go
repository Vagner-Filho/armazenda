package user_approval_service

import (
	"armazenda/entity/public"
	"armazenda/model/user_approval_model"
	"armazenda/service/user_service"
)

func GetPendingUsers(sessionId string) ([]entity_public.PendingUser, error) {
	farmId := user_service.GetFarmFromToken(sessionId)
	uam := user_approval_model.GetUserApprovalModel()
	return uam.GetPendingUsersByFarm(farmId)
}

func ApproveUser(userId uint32) error {
	uam := user_approval_model.GetUserApprovalModel()
	return uam.ApproveUser(userId)
}

func DeclineUser(userId uint32) error {
	uam := user_approval_model.GetUserApprovalModel()
	return uam.DeclineUser(userId)
}
