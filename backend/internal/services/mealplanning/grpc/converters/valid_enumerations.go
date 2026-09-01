package grpcconverters

import (
	"log"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	uploadedmediaconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	"github.com/primandproper/platform-go/v13/uploads/registry"
)

func ConvertGRPCCreateValidIngredientRequestToValidIngredientCreationRequestInput(request *mealplanninggrpc.ValidIngredientCreationRequestInput) *mealplanning.ValidIngredientCreationRequestInput {
	return &mealplanning.ValidIngredientCreationRequestInput{
		MinStorageTemperatureInCelsius: request.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: request.MaxStorageTemperatureInCelsius,
		Warning:                        request.Warning,
		IconPath:                       request.IconPath,
		PluralName:                     request.PluralName,
		StorageInstructions:            request.StorageInstructions,
		Name:                           request.Name,
		Description:                    request.Description,
		Slug:                           request.Slug,
		ShoppingSuggestions:            request.ShoppingSuggestions,
		ContainsFish:                   request.ContainsFish,
		ContainsShellfish:              request.ContainsShellfish,
		AnimalFlesh:                    request.AnimalFlesh,
		ContainsEgg:                    request.ContainsEgg,
		IsLiquid:                       request.IsLiquid,
		ContainsSoy:                    request.ContainsSoy,
		ContainsPeanut:                 request.ContainsPeanut,
		AnimalDerived:                  request.AnimalDerived,
		RestrictToPreparations:         request.RestrictToPreparations,
		ContaminatesEquipment:          request.ContaminatesEquipment,
		ContainsDairy:                  request.ContainsDairy,
		ContainsSesame:                 request.ContainsSesame,
		ContainsTreeNut:                request.ContainsTreeNut,
		ContainsWheat:                  request.ContainsWheat,
		ContainsAlcohol:                request.ContainsAlcohol,
		ContainsGluten:                 request.ContainsGluten,
		IsStarch:                       request.IsStarch,
		IsProtein:                      request.IsProtein,
		IsGrain:                        request.IsGrain,
		IsFruit:                        request.IsFruit,
		IsSalt:                         request.IsSalt,
		IsFat:                          request.IsFat,
		IsAcid:                         request.IsAcid,
		IsHeat:                         request.IsHeat,
	}
}

func ConvertGRPCValidIngredientUpdateRequestInputToValidIngredientUpdateRequestInput(x *mealplanninggrpc.ValidIngredientUpdateRequestInput) *mealplanning.ValidIngredientUpdateRequestInput {
	return &mealplanning.ValidIngredientUpdateRequestInput{
		Name:                           x.Name,
		Description:                    x.Description,
		Warning:                        x.Warning,
		IconPath:                       x.IconPath,
		ContainsDairy:                  x.ContainsDairy,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsEgg:                    x.ContainsEgg,
		ContainsWheat:                  x.ContainsWheat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		AnimalFlesh:                    x.AnimalFlesh,
		IsLiquid:                       x.IsLiquid,
		ContainsSoy:                    x.ContainsSoy,
		PluralName:                     x.PluralName,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Slug:                           x.Slug,
		ContainsAlcohol:                x.ContainsAlcohol,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
	}
}

func ConvertValidIngredientUpdateRequestInputToGRPCValidIngredientUpdateRequestInput(x *mealplanning.ValidIngredientUpdateRequestInput) *mealplanninggrpc.ValidIngredientUpdateRequestInput {
	return &mealplanninggrpc.ValidIngredientUpdateRequestInput{
		Name:                           x.Name,
		Description:                    x.Description,
		Warning:                        x.Warning,
		IconPath:                       x.IconPath,
		ContainsDairy:                  x.ContainsDairy,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsEgg:                    x.ContainsEgg,
		ContainsWheat:                  x.ContainsWheat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		AnimalFlesh:                    x.AnimalFlesh,
		IsLiquid:                       x.IsLiquid,
		ContainsSoy:                    x.ContainsSoy,
		PluralName:                     x.PluralName,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Slug:                           x.Slug,
		ContainsAlcohol:                x.ContainsAlcohol,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
	}
}

func ConvertValidIngredientToGRPCValidIngredient(x *mealplanning.ValidIngredient) *mealplanninggrpc.ValidIngredient {
	return &mealplanninggrpc.ValidIngredient{
		CreatedAt:                      converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Warning:                        x.Warning,
		PluralName:                     x.PluralName,
		IconPath:                       x.IconPath,
		Name:                           x.Name,
		Id:                             x.ID,
		Description:                    x.Description,
		Slug:                           x.Slug,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		ContainsEgg:                    x.ContainsEgg,
		ContainsAlcohol:                x.ContainsAlcohol,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsWheat:                  x.ContainsWheat,
		ContainsSoy:                    x.ContainsSoy,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		ContainsDairy:                  x.ContainsDairy,
		ContainsTreeNut:                x.ContainsTreeNut,
		AnimalFlesh:                    x.AnimalFlesh,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
		IsLiquid:                       x.IsLiquid,
		ContainsShellfish:              x.ContainsShellfish,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		Media:                          convertUploadedMediaSliceToGRPC(x.Media),
	}
}

func convertUploadedMediaSliceToGRPC(media []*registry.Object) []*uploadedmediasvc.UploadedMedia {
	if len(media) == 0 {
		return nil
	}
	out := make([]*uploadedmediasvc.UploadedMedia, len(media))
	for i, m := range media {
		out[i] = uploadedmediaconverters.ConvertUploadedMediaToGRPCUploadedMedia(m)
	}
	return out
}

func ConvertGRPCValidIngredientToValidIngredient(x *mealplanninggrpc.ValidIngredient) *mealplanning.ValidIngredient {
	return &mealplanning.ValidIngredient{
		CreatedAt:                      converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Warning:                        x.Warning,
		PluralName:                     x.PluralName,
		IconPath:                       x.IconPath,
		Name:                           x.Name,
		ID:                             x.Id,
		Description:                    x.Description,
		Slug:                           x.Slug,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		ContainsEgg:                    x.ContainsEgg,
		ContainsAlcohol:                x.ContainsAlcohol,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsWheat:                  x.ContainsWheat,
		ContainsSoy:                    x.ContainsSoy,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		ContainsDairy:                  x.ContainsDairy,
		ContainsTreeNut:                x.ContainsTreeNut,
		AnimalFlesh:                    x.AnimalFlesh,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
		IsLiquid:                       x.IsLiquid,
		ContainsShellfish:              x.ContainsShellfish,
		ContaminatesEquipment:          x.ContaminatesEquipment,
	}
}

func ConvertValidIngredientCreationRequestInputToGRPCValidIngredientCreationRequestInput(x *mealplanning.ValidIngredientCreationRequestInput) *mealplanninggrpc.ValidIngredientCreationRequestInput {
	return &mealplanninggrpc.ValidIngredientCreationRequestInput{
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		Warning:                        x.Warning,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		StorageInstructions:            x.StorageInstructions,
		Name:                           x.Name,
		Description:                    x.Description,
		Slug:                           x.Slug,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsAlcohol:                x.ContainsAlcohol,
		IsLiquid:                       x.IsLiquid,
		ContainsSoy:                    x.ContainsSoy,
		AnimalFlesh:                    x.AnimalFlesh,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		ContainsDairy:                  x.ContainsDairy,
		ContainsSesame:                 x.ContainsSesame,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsWheat:                  x.ContainsWheat,
		ContainsEgg:                    x.ContainsEgg,
		ContainsGluten:                 x.ContainsGluten,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsFish:                   x.ContainsFish,
	}
}

func ConvertGRPCValidIngredientGroupCreationRequestInputToValidIngredientGroupCreationRequestInput(request *mealplanninggrpc.ValidIngredientGroupCreationRequestInput) *mealplanning.ValidIngredientGroupCreationRequestInput {
	members := make([]*mealplanning.ValidIngredientGroupMemberCreationRequestInput, len(request.Members))
	for i, member := range request.Members {
		members[i] = &mealplanning.ValidIngredientGroupMemberCreationRequestInput{
			ValidIngredientID: member.ValidIngredientId,
		}
	}

	return &mealplanning.ValidIngredientGroupCreationRequestInput{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Members:     members,
	}
}

func ConvertValidIngredientGroupCreationRequestInputToGRPCValidIngredientGroupCreationRequestInput(request *mealplanning.ValidIngredientGroupCreationRequestInput) *mealplanninggrpc.ValidIngredientGroupCreationRequestInput {
	members := make([]*mealplanninggrpc.ValidIngredientGroupMemberCreationRequestInput, len(request.Members))
	for i, member := range request.Members {
		members[i] = &mealplanninggrpc.ValidIngredientGroupMemberCreationRequestInput{
			ValidIngredientId: member.ValidIngredientID,
		}
	}

	return &mealplanninggrpc.ValidIngredientGroupCreationRequestInput{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		Members:     members,
	}
}

func ConvertGRPCValidIngredientGroupUpdateRequestInputToValidIngredientGroupUpdateRequestInput(x *mealplanninggrpc.ValidIngredientGroupUpdateRequestInput) *mealplanning.ValidIngredientGroupUpdateRequestInput {
	return &mealplanning.ValidIngredientGroupUpdateRequestInput{
		Name:        x.Name,
		Slug:        x.Slug,
		Description: x.Description,
	}
}

func ConvertValidIngredientGroupUpdateRequestInputToGRPCValidIngredientGroupUpdateRequestInput(x *mealplanning.ValidIngredientGroupUpdateRequestInput) *mealplanninggrpc.ValidIngredientGroupUpdateRequestInput {
	return &mealplanninggrpc.ValidIngredientGroupUpdateRequestInput{
		Name:        x.Name,
		Slug:        x.Slug,
		Description: x.Description,
	}
}

func ConvertValidIngredientGroupToGRPCValidIngredientGroup(x *mealplanning.ValidIngredientGroup) *mealplanninggrpc.ValidIngredientGroup {
	members := make([]*mealplanninggrpc.ValidIngredientGroupMember, len(x.Members))
	for i, member := range x.Members {
		members[i] = &mealplanninggrpc.ValidIngredientGroupMember{
			CreatedAt:       converters.ConvertTimeToPBTimestamp(member.CreatedAt),
			ArchivedAt:      converters.ConvertTimePointerToPBTimestamp(member.ArchivedAt),
			Id:              member.ID,
			BelongsToGroup:  member.BelongsToGroup,
			ValidIngredient: ConvertValidIngredientToGRPCValidIngredient(&member.ValidIngredient),
		}
	}

	return &mealplanninggrpc.ValidIngredientGroup{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Id:            x.ID,
		Name:          x.Name,
		Slug:          x.Slug,
		Description:   x.Description,
		Members:       members,
	}
}

func ConvertGRPCValidIngredientGroupToValidIngredientGroup(x *mealplanninggrpc.ValidIngredientGroup) *mealplanning.ValidIngredientGroup {
	members := make([]*mealplanning.ValidIngredientGroupMember, len(x.Members))
	for i, member := range x.Members {
		members[i] = &mealplanning.ValidIngredientGroupMember{
			CreatedAt:       converters.ConvertPBTimestampToTime(member.CreatedAt),
			ArchivedAt:      converters.ConvertPBTimestampToTimePointer(member.ArchivedAt),
			ID:              member.Id,
			BelongsToGroup:  member.BelongsToGroup,
			ValidIngredient: *ConvertGRPCValidIngredientToValidIngredient(member.ValidIngredient),
		}
	}

	return &mealplanning.ValidIngredientGroup{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		ID:            x.Id,
		Name:          x.Name,
		Slug:          x.Slug,
		Description:   x.Description,
		Members:       members,
	}
}

func ConvertGRPCCreateValidIngredientMeasurementUnitRequestToValidIngredientMeasurementUnitCreationRequestInput(request *mealplanninggrpc.ValidIngredientMeasurementUnitCreationRequestInput) *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput {
	return &mealplanning.ValidIngredientMeasurementUnitCreationRequestInput{
		Notes:                  request.Notes,
		ValidMeasurementUnitID: request.ValidMeasurementUnitId,
		ValidIngredientID:      request.ValidIngredientId,
		MinAllowableQuantity:   request.MinAllowableQuantity,
		MaxAllowableQuantity:   request.MaxAllowableQuantity,
	}
}

func ConvertCreateValidIngredientMeasurementUnitRequestToGRPCValidIngredientMeasurementUnitCreationRequestInput(request *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput) *mealplanninggrpc.ValidIngredientMeasurementUnitCreationRequestInput {
	return &mealplanninggrpc.ValidIngredientMeasurementUnitCreationRequestInput{
		Notes:                  request.Notes,
		ValidMeasurementUnitId: request.ValidMeasurementUnitID,
		ValidIngredientId:      request.ValidIngredientID,
		MinAllowableQuantity:   request.MinAllowableQuantity,
		MaxAllowableQuantity:   request.MaxAllowableQuantity,
	}
}

func ConvertGRPCValidIngredientMeasurementUnitUpdateRequestInputToValidIngredientMeasurementUnitUpdateRequestInput(x *mealplanninggrpc.ValidIngredientMeasurementUnitUpdateRequestInput) *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput {
	return &mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput{
		Notes:                  x.Notes,
		ValidMeasurementUnitID: x.ValidMeasurementUnitId,
		ValidIngredientID:      x.ValidIngredientId,
		MinAllowableQuantity:   x.MinAllowableQuantity,
		MaxAllowableQuantity:   x.MaxAllowableQuantity,
	}
}

func ConvertValidIngredientMeasurementUnitUpdateRequestInputToGRPCValidIngredientMeasurementUnitUpdateRequestInput(x *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput) *mealplanninggrpc.ValidIngredientMeasurementUnitUpdateRequestInput {
	return &mealplanninggrpc.ValidIngredientMeasurementUnitUpdateRequestInput{
		Notes:                  x.Notes,
		ValidMeasurementUnitId: x.ValidMeasurementUnitID,
		ValidIngredientId:      x.ValidIngredientID,
		MinAllowableQuantity:   x.MinAllowableQuantity,
		MaxAllowableQuantity:   x.MaxAllowableQuantity,
	}
}

func ConvertValidIngredientMeasurementUnitToGRPCValidIngredientMeasurementUnit(x *mealplanning.ValidIngredientMeasurementUnit) *mealplanninggrpc.ValidIngredientMeasurementUnit {
	return &mealplanninggrpc.ValidIngredientMeasurementUnit{
		CreatedAt:            converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:        converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:           converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Notes:                x.Notes,
		Id:                   x.ID,
		MinAllowableQuantity: x.MinAllowableQuantity,
		MaxAllowableQuantity: x.MaxAllowableQuantity,
		MeasurementUnit:      ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.MeasurementUnit),
		Ingredient:           ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
	}
}

func ConvertGRPCValidIngredientMeasurementUnitToValidIngredientMeasurementUnit(x *mealplanninggrpc.ValidIngredientMeasurementUnit) *mealplanning.ValidIngredientMeasurementUnit {
	return &mealplanning.ValidIngredientMeasurementUnit{
		CreatedAt:            converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:        converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:           converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		Notes:                x.Notes,
		ID:                   x.Id,
		MinAllowableQuantity: x.MinAllowableQuantity,
		MaxAllowableQuantity: x.MaxAllowableQuantity,
		MeasurementUnit:      *ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(x.MeasurementUnit),
		Ingredient:           *ConvertGRPCValidIngredientToValidIngredient(x.Ingredient),
	}
}

func ConvertGRPCCreateValidIngredientPreparationRequestToValidIngredientPreparationCreationRequestInput(x *mealplanninggrpc.ValidIngredientPreparationCreationRequestInput) *mealplanning.ValidIngredientPreparationCreationRequestInput {
	return &mealplanning.ValidIngredientPreparationCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidIngredientID:  x.ValidIngredientId,
	}
}

func ConvertCreateValidIngredientPreparationRequestToGRPCValidIngredientPreparationCreationRequestInput(x *mealplanning.ValidIngredientPreparationCreationRequestInput) *mealplanninggrpc.ValidIngredientPreparationCreationRequestInput {
	return &mealplanninggrpc.ValidIngredientPreparationCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidIngredientId:  x.ValidIngredientID,
	}
}

func ConvertGRPCValidIngredientPreparationUpdateRequestInputToValidIngredientPreparationUpdateRequestInput(x *mealplanninggrpc.ValidIngredientPreparationUpdateRequestInput) *mealplanning.ValidIngredientPreparationUpdateRequestInput {
	return &mealplanning.ValidIngredientPreparationUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidIngredientID:  x.ValidIngredientId,
	}
}

func ConvertValidIngredientPreparationUpdateRequestInputToGRPCValidIngredientPreparationUpdateRequestInput(x *mealplanning.ValidIngredientPreparationUpdateRequestInput) *mealplanninggrpc.ValidIngredientPreparationUpdateRequestInput {
	return &mealplanninggrpc.ValidIngredientPreparationUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidIngredientId:  x.ValidIngredientID,
	}
}

func ConvertValidIngredientPreparationToGRPCValidIngredientPreparation(x *mealplanning.ValidIngredientPreparation) *mealplanninggrpc.ValidIngredientPreparation {
	return &mealplanninggrpc.ValidIngredientPreparation{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Notes:         x.Notes,
		Id:            x.ID,
		Preparation:   ConvertValidPreparationToGRPCValidPreparation(&x.Preparation),
		Ingredient:    ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
	}
}

func ConvertGRPCValidIngredientPreparationToValidIngredientPreparation(x *mealplanninggrpc.ValidIngredientPreparation) *mealplanning.ValidIngredientPreparation {
	return &mealplanning.ValidIngredientPreparation{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		Notes:         x.Notes,
		ID:            x.Id,
		Preparation:   *ConvertGRPCValidPreparationToValidPreparation(x.Preparation),
		Ingredient:    *ConvertGRPCValidIngredientToValidIngredient(x.Ingredient),
	}
}

func ConvertStringToValidIngredientStateAttributeType(s string) mealplanninggrpc.ValidIngredientStateAttributeType {
	switch s {
	case mealplanning.ValidIngredientStateAttributeTypeTexture:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TEXTURE
	case mealplanning.ValidIngredientStateAttributeTypeConsistency:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_CONSISTENCY
	case mealplanning.ValidIngredientStateAttributeTypeTemperature:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TEMPERATURE
	case mealplanning.ValidIngredientStateAttributeTypeColor:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_COLOR
	case mealplanning.ValidIngredientStateAttributeTypeAppearance:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_APPEARANCE
	case mealplanning.ValidIngredientStateAttributeTypeOdor:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_ODOR
	case mealplanning.ValidIngredientStateAttributeTypeTaste:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TASTE
	case mealplanning.ValidIngredientStateAttributeTypeSound:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_SOUND
	case mealplanning.ValidIngredientStateAttributeTypeOther:
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_OTHER
	default:
		log.Printf("UNKNOWN VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE: %q", s)
		return mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_OTHER
	}
}

func ConvertValidIngredientStateAttributeTypeToString(s mealplanninggrpc.ValidIngredientStateAttributeType) string {
	switch s {
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TEXTURE:
		return mealplanning.ValidIngredientStateAttributeTypeTexture
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_CONSISTENCY:
		return mealplanning.ValidIngredientStateAttributeTypeConsistency
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TEMPERATURE:
		return mealplanning.ValidIngredientStateAttributeTypeTemperature
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_COLOR:
		return mealplanning.ValidIngredientStateAttributeTypeColor
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_APPEARANCE:
		return mealplanning.ValidIngredientStateAttributeTypeAppearance
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_ODOR:
		return mealplanning.ValidIngredientStateAttributeTypeOdor
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_TASTE:
		return mealplanning.ValidIngredientStateAttributeTypeTaste
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_SOUND:
		return mealplanning.ValidIngredientStateAttributeTypeSound
	case mealplanninggrpc.ValidIngredientStateAttributeType_VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE_OTHER:
		return mealplanning.ValidIngredientStateAttributeTypeOther
	default:
		log.Printf("UNKNOWN VALID_INGREDIENT_STATE_ATTRIBUTE_TYPE: %q", s)
		return mealplanning.ValidIngredientStateAttributeTypeOther
	}
}

func ConvertGRPCCreateValidIngredientStateRequestToValidIngredientStateCreationRequestInput(x *mealplanninggrpc.ValidIngredientStateCreationRequestInput) *mealplanning.ValidIngredientStateCreationRequestInput {
	return &mealplanning.ValidIngredientStateCreationRequestInput{
		Name:          x.Name,
		Slug:          x.Slug,
		PastTense:     x.PastTense,
		Description:   x.Description,
		AttributeType: ConvertValidIngredientStateAttributeTypeToString(x.AttributeType),
		IconPath:      x.IconPath,
	}
}

func ConvertValidIngredientStateCreationRequestInputToGRPCValidIngredientStateCreationRequestInput(x *mealplanning.ValidIngredientStateCreationRequestInput) *mealplanninggrpc.ValidIngredientStateCreationRequestInput {
	return &mealplanninggrpc.ValidIngredientStateCreationRequestInput{
		Name:          x.Name,
		Slug:          x.Slug,
		PastTense:     x.PastTense,
		Description:   x.Description,
		AttributeType: ConvertStringToValidIngredientStateAttributeType(x.AttributeType),
		IconPath:      x.IconPath,
	}
}

func ConvertGRPCValidIngredientStateUpdateRequestInputToValidIngredientStateUpdateRequestInput(x *mealplanninggrpc.ValidIngredientStateUpdateRequestInput) *mealplanning.ValidIngredientStateUpdateRequestInput {
	var attributeType *string
	if x.AttributeType != nil {
		attributeType = new(ConvertValidIngredientStateAttributeTypeToString(*x.AttributeType))
	}

	return &mealplanning.ValidIngredientStateUpdateRequestInput{
		Name:          x.Name,
		Slug:          x.Slug,
		PastTense:     x.PastTense,
		Description:   x.Description,
		AttributeType: attributeType,
		IconPath:      x.IconPath,
	}
}

func ConvertValidIngredientStateUpdateRequestInputToGRPCValidIngredientStateUpdateRequestInput(x *mealplanning.ValidIngredientStateUpdateRequestInput) *mealplanninggrpc.ValidIngredientStateUpdateRequestInput {
	var attributeType *mealplanninggrpc.ValidIngredientStateAttributeType
	if x.AttributeType != nil {
		attributeType = new(ConvertStringToValidIngredientStateAttributeType(*x.AttributeType))
	}

	return &mealplanninggrpc.ValidIngredientStateUpdateRequestInput{
		Name:          x.Name,
		Slug:          x.Slug,
		PastTense:     x.PastTense,
		Description:   x.Description,
		AttributeType: attributeType,
		IconPath:      x.IconPath,
	}
}

func ConvertValidIngredientStateToGRPCValidIngredientState(x *mealplanning.ValidIngredientState) *mealplanninggrpc.ValidIngredientState {
	return &mealplanninggrpc.ValidIngredientState{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		PastTense:     x.PastTense,
		Description:   x.Description,
		IconPath:      x.IconPath,
		Id:            x.ID,
		Name:          x.Name,
		AttributeType: ConvertStringToValidIngredientStateAttributeType(x.AttributeType),
		Slug:          x.Slug,
	}
}

func ConvertGRPCValidIngredientStateToValidIngredientState(x *mealplanninggrpc.ValidIngredientState) *mealplanning.ValidIngredientState {
	return &mealplanning.ValidIngredientState{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		PastTense:     x.PastTense,
		Description:   x.Description,
		IconPath:      x.IconPath,
		ID:            x.Id,
		Name:          x.Name,
		AttributeType: ConvertValidIngredientStateAttributeTypeToString(x.AttributeType),
		Slug:          x.Slug,
	}
}

func ConvertGRPCCreateValidIngredientStateIngredientRequestToValidIngredientStateIngredientCreationRequestInput(x *mealplanninggrpc.ValidIngredientStateIngredientCreationRequestInput) *mealplanning.ValidIngredientStateIngredientCreationRequestInput {
	return &mealplanning.ValidIngredientStateIngredientCreationRequestInput{
		Notes:                  x.Notes,
		ValidIngredientStateID: x.ValidIngredientStateId,
		ValidIngredientID:      x.ValidIngredientId,
	}
}

func ConvertCreateValidIngredientStateIngredientRequestToGRPCValidIngredientStateIngredientCreationRequestInput(x *mealplanning.ValidIngredientStateIngredientCreationRequestInput) *mealplanninggrpc.ValidIngredientStateIngredientCreationRequestInput {
	return &mealplanninggrpc.ValidIngredientStateIngredientCreationRequestInput{
		Notes:                  x.Notes,
		ValidIngredientStateId: x.ValidIngredientStateID,
		ValidIngredientId:      x.ValidIngredientID,
	}
}

func ConvertGRPCValidIngredientStateIngredientUpdateRequestInputToValidIngredientStateIngredientUpdateRequestInput(x *mealplanninggrpc.ValidIngredientStateIngredientUpdateRequestInput) *mealplanning.ValidIngredientStateIngredientUpdateRequestInput {
	return &mealplanning.ValidIngredientStateIngredientUpdateRequestInput{
		Notes:                  x.Notes,
		ValidIngredientStateID: x.ValidIngredientStateId,
		ValidIngredientID:      x.ValidIngredientId,
	}
}

func ConvertValidIngredientStateIngredientUpdateRequestInputToGRPCValidIngredientStateIngredientUpdateRequestInput(x *mealplanning.ValidIngredientStateIngredientUpdateRequestInput) *mealplanninggrpc.ValidIngredientStateIngredientUpdateRequestInput {
	return &mealplanninggrpc.ValidIngredientStateIngredientUpdateRequestInput{
		Notes:                  x.Notes,
		ValidIngredientStateId: x.ValidIngredientStateID,
		ValidIngredientId:      x.ValidIngredientID,
	}
}

func ConvertValidIngredientStateIngredientToGRPCValidIngredientStateIngredient(x *mealplanning.ValidIngredientStateIngredient) *mealplanninggrpc.ValidIngredientStateIngredient {
	return &mealplanninggrpc.ValidIngredientStateIngredient{
		CreatedAt:       converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:   converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:      converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Notes:           x.Notes,
		Id:              x.ID,
		IngredientState: ConvertValidIngredientStateToGRPCValidIngredientState(&x.IngredientState),
		Ingredient:      ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
	}
}

func ConvertGRPCValidIngredientStateIngredientToValidIngredientStateIngredient(x *mealplanninggrpc.ValidIngredientStateIngredient) *mealplanning.ValidIngredientStateIngredient {
	return &mealplanning.ValidIngredientStateIngredient{
		CreatedAt:       converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:   converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:      converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		Notes:           x.Notes,
		ID:              x.Id,
		IngredientState: *ConvertGRPCValidIngredientStateToValidIngredientState(x.IngredientState),
		Ingredient:      *ConvertGRPCValidIngredientToValidIngredient(x.Ingredient),
	}
}

func ConvertGRPCCreateValidInstrumentRequestToValidInstrumentCreationRequestInput(x *mealplanninggrpc.ValidInstrumentCreationRequestInput) *mealplanning.ValidInstrumentCreationRequestInput {
	return &mealplanning.ValidInstrumentCreationRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertGRPCValidInstrumentUpdateRequestInputToValidInstrumentUpdateRequestInput(x *mealplanninggrpc.ValidInstrumentUpdateRequestInput) *mealplanning.ValidInstrumentUpdateRequestInput {
	return &mealplanning.ValidInstrumentUpdateRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		Slug:                           x.Slug,
		UsableForStorage:               x.UsableForStorage,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
	}
}

func ConvertValidInstrumentUpdateRequestInputToGRPCValidInstrumentUpdateRequestInput(x *mealplanning.ValidInstrumentUpdateRequestInput) *mealplanninggrpc.ValidInstrumentUpdateRequestInput {
	return &mealplanninggrpc.ValidInstrumentUpdateRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		Slug:                           x.Slug,
		UsableForStorage:               x.UsableForStorage,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
	}
}

func ConvertValidInstrumentToGRPCValidInstrument(x *mealplanning.ValidInstrument) *mealplanninggrpc.ValidInstrument {
	return &mealplanninggrpc.ValidInstrument{
		CreatedAt:                      converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Name:                           x.Name,
		Id:                             x.ID,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertGRPCValidInstrumentToValidInstrument(x *mealplanninggrpc.ValidInstrument) *mealplanning.ValidInstrument {
	return &mealplanning.ValidInstrument{
		CreatedAt:                      converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		Name:                           x.Name,
		ID:                             x.Id,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertValidInstrumentCreationRequestInputToGRPCValidInstrumentCreationRequestInput(x *mealplanning.ValidInstrumentCreationRequestInput) *mealplanninggrpc.ValidInstrumentCreationRequestInput {
	return &mealplanninggrpc.ValidInstrumentCreationRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertGRPCValidMeasurementUnitCreationRequestInputToValidMeasurementUnitCreationRequestInput(x *mealplanninggrpc.ValidMeasurementUnitCreationRequestInput) *mealplanning.ValidMeasurementUnitCreationRequestInput {
	return &mealplanning.ValidMeasurementUnitCreationRequestInput{
		Name:        x.Name,
		Description: x.Description,
		IconPath:    x.IconPath,
		PluralName:  x.PluralName,
		Slug:        x.Slug,
		Volumetric:  x.Volumetric,
		Universal:   x.Universal,
		Metric:      x.Metric,
		Imperial:    x.Imperial,
	}
}

func ConvertValidMeasurementUnitCreationRequestInputToGRPCValidMeasurementUnitCreationRequestInput(x *mealplanning.ValidMeasurementUnitCreationRequestInput) *mealplanninggrpc.ValidMeasurementUnitCreationRequestInput {
	return &mealplanninggrpc.ValidMeasurementUnitCreationRequestInput{
		Name:        x.Name,
		Description: x.Description,
		IconPath:    x.IconPath,
		PluralName:  x.PluralName,
		Slug:        x.Slug,
		Volumetric:  x.Volumetric,
		Universal:   x.Universal,
		Metric:      x.Metric,
		Imperial:    x.Imperial,
	}
}

func ConvertGRPCValidMeasurementUnitUpdateRequestInputToValidMeasurementUnitUpdateRequestInput(x *mealplanninggrpc.ValidMeasurementUnitUpdateRequestInput) *mealplanning.ValidMeasurementUnitUpdateRequestInput {
	return &mealplanning.ValidMeasurementUnitUpdateRequestInput{
		Name:        x.Name,
		Description: x.Description,
		IconPath:    x.IconPath,
		Volumetric:  x.Volumetric,
		Universal:   x.Universal,
		Metric:      x.Metric,
		Imperial:    x.Imperial,
		PluralName:  x.PluralName,
		Slug:        x.Slug,
	}
}

func ConvertValidMeasurementUnitUpdateRequestInputToGRPCValidMeasurementUnitUpdateRequestInput(x *mealplanning.ValidMeasurementUnitUpdateRequestInput) *mealplanninggrpc.ValidMeasurementUnitUpdateRequestInput {
	return &mealplanninggrpc.ValidMeasurementUnitUpdateRequestInput{
		Name:        x.Name,
		Description: x.Description,
		IconPath:    x.IconPath,
		Volumetric:  x.Volumetric,
		Universal:   x.Universal,
		Metric:      x.Metric,
		Imperial:    x.Imperial,
		PluralName:  x.PluralName,
		Slug:        x.Slug,
	}
}

func ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(x *mealplanning.ValidMeasurementUnit) *mealplanninggrpc.ValidMeasurementUnit {
	return &mealplanninggrpc.ValidMeasurementUnit{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		PluralName:    x.PluralName,
		IconPath:      x.IconPath,
		Id:            x.ID,
		Description:   x.Description,
		Name:          x.Name,
		Slug:          x.Slug,
		Volumetric:    x.Volumetric,
		Universal:     x.Universal,
		Metric:        x.Metric,
		Imperial:      x.Imperial,
	}
}

func ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(x *mealplanninggrpc.ValidMeasurementUnit) *mealplanning.ValidMeasurementUnit {
	return &mealplanning.ValidMeasurementUnit{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		PluralName:    x.PluralName,
		IconPath:      x.IconPath,
		ID:            x.Id,
		Description:   x.Description,
		Name:          x.Name,
		Slug:          x.Slug,
		Volumetric:    x.Volumetric,
		Universal:     x.Universal,
		Metric:        x.Metric,
		Imperial:      x.Imperial,
	}
}

func ConvertGRPCCreateValidMeasurementUnitConversionRequestToValidMeasurementUnitConversionCreationRequestInput(x *mealplanninggrpc.ValidMeasurementUnitConversionCreationRequestInput) *mealplanning.ValidMeasurementUnitConversionCreationRequestInput {
	return &mealplanning.ValidMeasurementUnitConversionCreationRequestInput{
		OnlyForIngredient: x.OnlyForIngredient,
		From:              x.From,
		To:                x.To,
		Notes:             x.Notes,
		Modifier:          x.Modifier,
	}
}

func ConvertCreateValidMeasurementUnitConversionRequestToGRPCValidMeasurementUnitConversionCreationRequestInput(x *mealplanning.ValidMeasurementUnitConversionCreationRequestInput) *mealplanninggrpc.ValidMeasurementUnitConversionCreationRequestInput {
	return &mealplanninggrpc.ValidMeasurementUnitConversionCreationRequestInput{
		OnlyForIngredient: x.OnlyForIngredient,
		From:              x.From,
		To:                x.To,
		Notes:             x.Notes,
		Modifier:          x.Modifier,
	}
}

func ConvertGRPCValidMeasurementUnitConversionUpdateRequestInputToValidMeasurementUnitConversionUpdateRequestInput(x *mealplanninggrpc.ValidMeasurementUnitConversionUpdateRequestInput) *mealplanning.ValidMeasurementUnitConversionUpdateRequestInput {
	return &mealplanning.ValidMeasurementUnitConversionUpdateRequestInput{
		OnlyForIngredient: x.OnlyForIngredient,
		From:              x.From,
		To:                x.To,
		Notes:             x.Notes,
		Modifier:          x.Modifier,
	}
}

func ConvertValidMeasurementUnitConversionUpdateRequestInputToGRPCValidMeasurementUnitConversionUpdateRequestInput(x *mealplanning.ValidMeasurementUnitConversionUpdateRequestInput) *mealplanninggrpc.ValidMeasurementUnitConversionUpdateRequestInput {
	return &mealplanninggrpc.ValidMeasurementUnitConversionUpdateRequestInput{
		OnlyForIngredient: x.OnlyForIngredient,
		From:              x.From,
		To:                x.To,
		Notes:             x.Notes,
		Modifier:          x.Modifier,
	}
}

func ConvertValidMeasurementUnitConversionToGRPCValidMeasurementUnitConversion(x *mealplanning.ValidMeasurementUnitConversion) *mealplanninggrpc.ValidMeasurementUnitConversion {
	var ingredient *mealplanninggrpc.ValidIngredient
	if x.OnlyForIngredient != nil {
		ingredient = ConvertValidIngredientToGRPCValidIngredient(x.OnlyForIngredient)
	}

	y := &mealplanninggrpc.ValidMeasurementUnitConversion{
		CreatedAt:         converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:     converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:        converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		From:              ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.From),
		To:                ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.To),
		Notes:             x.Notes,
		OnlyForIngredient: ingredient,
		Id:                x.ID,
		Modifier:          x.Modifier,
	}

	return y
}

func ConvertMeasurementUnitConversionMismatchToGRPCMeasurementUnitConversionMismatch(x *mealplanning.MeasurementUnitConversionMismatch) *mealplanninggrpc.MeasurementUnitConversionMismatch {
	if x == nil {
		return nil
	}
	return &mealplanninggrpc.MeasurementUnitConversionMismatch{
		Ingredient: ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
		FromUnit:   ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.FromUnit),
		ToUnit:     ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&x.ToUnit),
	}
}

func ConvertGRPCValidMeasurementUnitConversionToValidMeasurementUnitConversion(x *mealplanninggrpc.ValidMeasurementUnitConversion) *mealplanning.ValidMeasurementUnitConversion {
	var ingredient *mealplanning.ValidIngredient
	if x.OnlyForIngredient != nil {
		ingredient = ConvertGRPCValidIngredientToValidIngredient(x.OnlyForIngredient)
	}

	y := &mealplanning.ValidMeasurementUnitConversion{
		CreatedAt:         converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:     converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:        converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		From:              *ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(x.From),
		To:                *ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(x.To),
		Notes:             x.Notes,
		OnlyForIngredient: ingredient,
		ID:                x.Id,
		Modifier:          x.Modifier,
	}

	return y
}

func ConvertGRPCValidPreparationCreationRequestInputToValidPreparationCreationRequestInput(x *mealplanninggrpc.ValidPreparationCreationRequestInput) *mealplanning.ValidPreparationCreationRequestInput {
	return &mealplanning.ValidPreparationCreationRequestInput{
		MinInstrumentCount:          uint16(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          uint16(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxIngredientCount),
		MinVesselCount:              uint16(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint32PointerToUint16Pointer(x.MaxVesselCount),
		IconPath:                    x.IconPath,
		PastTense:                   x.PastTense,
		Slug:                        x.Slug,
		Name:                        x.Name,
		Description:                 x.Description,
		TemperatureRequired:         x.TemperatureRequired,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
		RestrictToIngredients:       x.RestrictToIngredients,
		YieldsNothing:               x.YieldsNothing,
	}
}

func ConvertValidPreparationCreationRequestInputToGRPCValidPreparationCreationRequestInput(x *mealplanning.ValidPreparationCreationRequestInput) *mealplanninggrpc.ValidPreparationCreationRequestInput {
	return &mealplanninggrpc.ValidPreparationCreationRequestInput{
		MinInstrumentCount:          uint32(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          uint32(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxIngredientCount),
		MinVesselCount:              uint32(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint16PointerToUint32Pointer(x.MaxVesselCount),
		IconPath:                    x.IconPath,
		PastTense:                   x.PastTense,
		Slug:                        x.Slug,
		Name:                        x.Name,
		Description:                 x.Description,
		TemperatureRequired:         x.TemperatureRequired,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
		RestrictToIngredients:       x.RestrictToIngredients,
		YieldsNothing:               x.YieldsNothing,
	}
}

func ConvertGRPCValidPreparationUpdateRequestInputToValidPreparationUpdateRequestInput(x *mealplanninggrpc.ValidPreparationUpdateRequestInput) *mealplanning.ValidPreparationUpdateRequestInput {
	return &mealplanning.ValidPreparationUpdateRequestInput{
		MinInstrumentCount:          converters.ConvertUint32PointerToUint16Pointer(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          converters.ConvertUint32PointerToUint16Pointer(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxIngredientCount),
		MinVesselCount:              converters.ConvertUint32PointerToUint16Pointer(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint32PointerToUint16Pointer(x.MaxVesselCount),
		Name:                        x.Name,
		Description:                 x.Description,
		IconPath:                    x.IconPath,
		YieldsNothing:               x.YieldsNothing,
		Slug:                        x.Slug,
		RestrictToIngredients:       x.RestrictToIngredients,
		PastTense:                   x.PastTense,
		TemperatureRequired:         x.TemperatureRequired,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
	}
}

func ConvertValidPreparationUpdateRequestInputToGRPCValidPreparationUpdateRequestInput(x *mealplanning.ValidPreparationUpdateRequestInput) *mealplanninggrpc.ValidPreparationUpdateRequestInput {
	return &mealplanninggrpc.ValidPreparationUpdateRequestInput{
		MinInstrumentCount:          converters.ConvertUint16PointerToUint32Pointer(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          converters.ConvertUint16PointerToUint32Pointer(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxIngredientCount),
		MinVesselCount:              converters.ConvertUint16PointerToUint32Pointer(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint16PointerToUint32Pointer(x.MaxVesselCount),
		Name:                        x.Name,
		Description:                 x.Description,
		IconPath:                    x.IconPath,
		YieldsNothing:               x.YieldsNothing,
		Slug:                        x.Slug,
		RestrictToIngredients:       x.RestrictToIngredients,
		PastTense:                   x.PastTense,
		TemperatureRequired:         x.TemperatureRequired,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
	}
}

func ConvertValidPreparationToGRPCValidPreparation(x *mealplanning.ValidPreparation) *mealplanninggrpc.ValidPreparation {
	return &mealplanninggrpc.ValidPreparation{
		CreatedAt:                   converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:               converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:                  converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		MinInstrumentCount:          uint32(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          uint32(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint16PointerToUint32Pointer(x.MaxIngredientCount),
		MinVesselCount:              uint32(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint16PointerToUint32Pointer(x.MaxVesselCount),
		Name:                        x.Name,
		Id:                          x.ID,
		IconPath:                    x.IconPath,
		Description:                 x.Description,
		Slug:                        x.Slug,
		PastTense:                   x.PastTense,
		TemperatureRequired:         x.TemperatureRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
		YieldsNothing:               x.YieldsNothing,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		RestrictToIngredients:       x.RestrictToIngredients,
		Media:                       convertUploadedMediaSliceToGRPC(x.Media),
	}
}

func ConvertGRPCValidPreparationToValidPreparation(x *mealplanninggrpc.ValidPreparation) *mealplanning.ValidPreparation {
	return &mealplanning.ValidPreparation{
		MinInstrumentCount:          uint16(x.MinInstrumentCount),
		MaxInstrumentCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxInstrumentCount),
		MinIngredientCount:          uint16(x.MinIngredientCount),
		MaxIngredientCount:          converters.ConvertUint32PointerToUint16Pointer(x.MaxIngredientCount),
		MinVesselCount:              uint16(x.MinVesselCount),
		MaxVesselCount:              converters.ConvertUint32PointerToUint16Pointer(x.MaxVesselCount),
		CreatedAt:                   converters.ConvertPBTimestampToTime(x.CreatedAt),
		ArchivedAt:                  converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		LastUpdatedAt:               converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		IconPath:                    x.IconPath,
		PastTense:                   x.PastTense,
		ID:                          x.Id,
		Name:                        x.Name,
		Description:                 x.Description,
		Slug:                        x.Slug,
		RestrictToIngredients:       x.RestrictToIngredients,
		TemperatureRequired:         x.TemperatureRequired,
		TimeEstimateRequired:        x.TimeEstimateRequired,
		ConditionExpressionRequired: x.ConditionExpressionRequired,
		ConsumesVessel:              x.ConsumesVessel,
		OnlyForVessels:              x.OnlyForVessels,
		YieldsNothing:               x.YieldsNothing,
	}
}

func ConvertGRPCCreateValidPreparationInstrumentRequestToValidPreparationInstrumentCreationRequestInput(x *mealplanninggrpc.ValidPreparationInstrumentCreationRequestInput) *mealplanning.ValidPreparationInstrumentCreationRequestInput {
	return &mealplanning.ValidPreparationInstrumentCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidInstrumentID:  x.ValidInstrumentId,
	}
}

func ConvertCreateValidPreparationInstrumentRequestToGRPCValidPreparationInstrumentCreationRequestInput(x *mealplanning.ValidPreparationInstrumentCreationRequestInput) *mealplanninggrpc.ValidPreparationInstrumentCreationRequestInput {
	return &mealplanninggrpc.ValidPreparationInstrumentCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidInstrumentId:  x.ValidInstrumentID,
	}
}

func ConvertGRPCValidPreparationInstrumentUpdateRequestInputToValidPreparationInstrumentUpdateRequestInput(x *mealplanninggrpc.ValidPreparationInstrumentUpdateRequestInput) *mealplanning.ValidPreparationInstrumentUpdateRequestInput {
	return &mealplanning.ValidPreparationInstrumentUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidInstrumentID:  x.ValidInstrumentId,
	}
}

func ConvertValidPreparationInstrumentUpdateRequestInputToGRPCValidPreparationInstrumentUpdateRequestInput(x *mealplanning.ValidPreparationInstrumentUpdateRequestInput) *mealplanninggrpc.ValidPreparationInstrumentUpdateRequestInput {
	return &mealplanninggrpc.ValidPreparationInstrumentUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidInstrumentId:  x.ValidInstrumentID,
	}
}

func ConvertValidPreparationInstrumentToGRPCValidPreparationInstrument(x *mealplanning.ValidPreparationInstrument) *mealplanninggrpc.ValidPreparationInstrument {
	return &mealplanninggrpc.ValidPreparationInstrument{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Id:            x.ID,
		Notes:         x.Notes,
		Instrument:    ConvertValidInstrumentToGRPCValidInstrument(&x.Instrument),
		Preparation:   ConvertValidPreparationToGRPCValidPreparation(&x.Preparation),
	}
}

func ConvertGRPCValidPreparationInstrumentToValidPreparationInstrument(x *mealplanninggrpc.ValidPreparationInstrument) *mealplanning.ValidPreparationInstrument {
	return &mealplanning.ValidPreparationInstrument{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		ID:            x.Id,
		Notes:         x.Notes,
		Instrument:    *ConvertGRPCValidInstrumentToValidInstrument(x.Instrument),
		Preparation:   *ConvertGRPCValidPreparationToValidPreparation(x.Preparation),
	}
}

func ConvertGRPCCreateValidPreparationVesselRequestToValidPreparationVesselCreationRequestInput(x *mealplanninggrpc.ValidPreparationVesselCreationRequestInput) *mealplanning.ValidPreparationVesselCreationRequestInput {
	return &mealplanning.ValidPreparationVesselCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidVesselID:      x.ValidVesselId,
	}
}

func ConvertCreateValidPreparationVesselRequestToGRPCValidPreparationVesselCreationRequestInput(x *mealplanning.ValidPreparationVesselCreationRequestInput) *mealplanninggrpc.ValidPreparationVesselCreationRequestInput {
	return &mealplanninggrpc.ValidPreparationVesselCreationRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidVesselId:      x.ValidVesselID,
	}
}

func ConvertGRPCValidPreparationVesselUpdateRequestInputToValidPreparationVesselUpdateRequestInput(x *mealplanninggrpc.ValidPreparationVesselUpdateRequestInput) *mealplanning.ValidPreparationVesselUpdateRequestInput {
	return &mealplanning.ValidPreparationVesselUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationID: x.ValidPreparationId,
		ValidVesselID:      x.ValidVesselId,
	}
}

func ConvertValidPreparationVesselUpdateRequestInputToGRPCValidPreparationVesselUpdateRequestInput(x *mealplanning.ValidPreparationVesselUpdateRequestInput) *mealplanninggrpc.ValidPreparationVesselUpdateRequestInput {
	return &mealplanninggrpc.ValidPreparationVesselUpdateRequestInput{
		Notes:              x.Notes,
		ValidPreparationId: x.ValidPreparationID,
		ValidVesselId:      x.ValidVesselID,
	}
}

func ConvertValidPreparationVesselToGRPCValidPreparationVessel(x *mealplanning.ValidPreparationVessel) *mealplanninggrpc.ValidPreparationVessel {
	return &mealplanninggrpc.ValidPreparationVessel{
		CreatedAt:     converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt: converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		Id:            x.ID,
		Notes:         x.Notes,
		Preparation:   ConvertValidPreparationToGRPCValidPreparation(&x.Preparation),
		Vessel:        ConvertValidVesselToGRPCValidVessel(&x.Vessel),
	}
}

func ConvertGRPCValidPreparationVesselToValidPreparationVessel(x *mealplanninggrpc.ValidPreparationVessel) *mealplanning.ValidPreparationVessel {
	return &mealplanning.ValidPreparationVessel{
		CreatedAt:     converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt: converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:    converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		ID:            x.Id,
		Notes:         x.Notes,
		Preparation:   *ConvertGRPCValidPreparationToValidPreparation(x.Preparation),
		Vessel:        *ConvertGRPCValidVesselToValidVessel(x.Vessel),
	}
}

func ConvertStringToValidVesselShape(s string) mealplanninggrpc.ValidVesselShape {
	switch s {
	case mealplanning.VesselShapeHemisphere:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_HEMISPHERE
	case mealplanning.VesselShapeRectangle:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_RECTANGLE
	case mealplanning.VesselShapeCone:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CONE
	case mealplanning.VesselShapePyramid:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_PYRAMID
	case mealplanning.VesselShapeCylinder:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CYLINDER
	case mealplanning.VesselShapeSphere:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_SPHERE
	case mealplanning.VesselShapeCube:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CUBE
	case mealplanning.VesselShapeOther:
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_OTHER
	default:
		log.Printf("UNKNOWN VESSEL SHAPE: %q", s)
		return mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_OTHER
	}
}
func ConvertValidVesselShapeToString(s mealplanninggrpc.ValidVesselShape) string {
	switch s {
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_HEMISPHERE:
		return mealplanning.VesselShapeHemisphere
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_RECTANGLE:
		return mealplanning.VesselShapeRectangle
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CONE:
		return mealplanning.VesselShapeCone
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_PYRAMID:
		return mealplanning.VesselShapePyramid
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CYLINDER:
		return mealplanning.VesselShapeCylinder
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_SPHERE:
		return mealplanning.VesselShapeSphere
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_CUBE:
		return mealplanning.VesselShapeCube
	case mealplanninggrpc.ValidVesselShape_VESSEL_SHAPE_OTHER:
		return mealplanning.VesselShapeOther
	default:
		log.Printf("UNKNOWN VESSEL SHAPE: %q", s)
		return mealplanning.VesselShapeOther
	}
}

func ConvertGRPCValidVesselCreationRequestInputToValidVesselCreationRequestInput(x *mealplanninggrpc.ValidVesselCreationRequestInput) *mealplanning.ValidVesselCreationRequestInput {
	return &mealplanning.ValidVesselCreationRequestInput{
		CapacityUnitID:                 x.CapacityUnitId,
		Shape:                          ConvertValidVesselShapeToString(x.Shape),
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		Name:                           x.Name,
		Description:                    x.Description,
		Slug:                           x.Slug,
		LengthInMillimeters:            x.LengthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Capacity:                       x.Capacity,
		WidthInMillimeters:             x.WidthInMillimeters,
		UsableForStorage:               x.UsableForStorage,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
	}
}

func ConvertValidVesselCreationRequestInputToGRPCValidVesselCreationRequestInput(x *mealplanning.ValidVesselCreationRequestInput) *mealplanninggrpc.ValidVesselCreationRequestInput {
	return &mealplanninggrpc.ValidVesselCreationRequestInput{
		CapacityUnitId:                 x.CapacityUnitID,
		Shape:                          ConvertStringToValidVesselShape(x.Shape),
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		Name:                           x.Name,
		Description:                    x.Description,
		Slug:                           x.Slug,
		LengthInMillimeters:            x.LengthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Capacity:                       x.Capacity,
		WidthInMillimeters:             x.WidthInMillimeters,
		UsableForStorage:               x.UsableForStorage,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
	}
}

func ConvertGRPCValidVesselUpdateRequestInputToValidVesselUpdateRequestInput(x *mealplanninggrpc.ValidVesselUpdateRequestInput) *mealplanning.ValidVesselUpdateRequestInput {
	var shape *string
	if x.Shape != nil {
		shape = new(ConvertValidVesselShapeToString(*x.Shape))
	}

	return &mealplanning.ValidVesselUpdateRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		UsableForStorage:               x.UsableForStorage,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		Capacity:                       x.Capacity,
		CapacityUnitID:                 x.CapacityUnitId,
		WidthInMillimeters:             x.WidthInMillimeters,
		LengthInMillimeters:            x.LengthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Shape:                          shape,
	}
}

func ConvertValidVesselUpdateRequestInputToGRPCValidVesselUpdateRequestInput(x *mealplanning.ValidVesselUpdateRequestInput) *mealplanninggrpc.ValidVesselUpdateRequestInput {
	var shape *mealplanninggrpc.ValidVesselShape
	if x.Shape != nil {
		shape = new(ConvertStringToValidVesselShape(*x.Shape))
	}

	return &mealplanninggrpc.ValidVesselUpdateRequestInput{
		Name:                           x.Name,
		PluralName:                     x.PluralName,
		Description:                    x.Description,
		IconPath:                       x.IconPath,
		UsableForStorage:               x.UsableForStorage,
		Slug:                           x.Slug,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		Capacity:                       x.Capacity,
		CapacityUnitId:                 x.CapacityUnitID,
		WidthInMillimeters:             x.WidthInMillimeters,
		LengthInMillimeters:            x.LengthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Shape:                          shape,
	}
}

func ConvertValidVesselToGRPCValidVessel(x *mealplanning.ValidVessel) *mealplanninggrpc.ValidVessel {
	var capacityUnit *mealplanninggrpc.ValidMeasurementUnit
	if x.CapacityUnit != nil {
		capacityUnit = ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(x.CapacityUnit)
	}

	return &mealplanninggrpc.ValidVessel{
		CreatedAt:                      converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		CapacityUnit:                   capacityUnit,
		Shape:                          ConvertStringToValidVesselShape(x.Shape),
		Description:                    x.Description,
		Name:                           x.Name,
		Slug:                           x.Slug,
		IconPath:                       x.IconPath,
		Id:                             x.ID,
		PluralName:                     x.PluralName,
		WidthInMillimeters:             x.WidthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Capacity:                       x.Capacity,
		LengthInMillimeters:            x.LengthInMillimeters,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertGRPCValidVesselToValidVessel(x *mealplanninggrpc.ValidVessel) *mealplanning.ValidVessel {
	var capacityUnit *mealplanning.ValidMeasurementUnit
	if x.CapacityUnit != nil {
		capacityUnit = ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(x.CapacityUnit)
	}

	return &mealplanning.ValidVessel{
		CreatedAt:                      converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		CapacityUnit:                   capacityUnit,
		Shape:                          ConvertValidVesselShapeToString(x.Shape),
		Description:                    x.Description,
		Name:                           x.Name,
		Slug:                           x.Slug,
		IconPath:                       x.IconPath,
		ID:                             x.Id,
		PluralName:                     x.PluralName,
		WidthInMillimeters:             x.WidthInMillimeters,
		HeightInMillimeters:            x.HeightInMillimeters,
		Capacity:                       x.Capacity,
		LengthInMillimeters:            x.LengthInMillimeters,
		IncludeInGeneratedInstructions: x.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          x.DisplayInSummaryLists,
		UsableForStorage:               x.UsableForStorage,
	}
}

func ConvertGRPCValidPrepTaskConfigCreationRequestInputToValidPrepTaskConfigCreationRequestInput(x *mealplanninggrpc.ValidPrepTaskConfigCreationRequestInput) *mealplanning.ValidPrepTaskConfigCreationRequestInput {
	return &mealplanning.ValidPrepTaskConfigCreationRequestInput{
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		ValidPreparationID:             x.ValidPreparationId,
		ValidIngredientID:              x.ValidIngredientId,
	}
}

func ConvertValidPrepTaskConfigCreationRequestInputToGRPCValidPrepTaskConfigCreationRequestInput(x *mealplanning.ValidPrepTaskConfigCreationRequestInput) *mealplanninggrpc.ValidPrepTaskConfigCreationRequestInput {
	return &mealplanninggrpc.ValidPrepTaskConfigCreationRequestInput{
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		ValidPreparationId:             x.ValidPreparationID,
		ValidIngredientId:              x.ValidIngredientID,
	}
}

func ConvertGRPCValidPrepTaskConfigUpdateRequestInputToValidPrepTaskConfigUpdateRequestInput(x *mealplanninggrpc.ValidPrepTaskConfigUpdateRequestInput) *mealplanning.ValidPrepTaskConfigUpdateRequestInput {
	return &mealplanning.ValidPrepTaskConfigUpdateRequestInput{
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		ValidPreparationID:             x.ValidPreparationId,
		ValidIngredientID:              x.ValidIngredientId,
	}
}

func ConvertValidPrepTaskConfigUpdateRequestInputToGRPCValidPrepTaskConfigUpdateRequestInput(x *mealplanning.ValidPrepTaskConfigUpdateRequestInput) *mealplanninggrpc.ValidPrepTaskConfigUpdateRequestInput {
	return &mealplanninggrpc.ValidPrepTaskConfigUpdateRequestInput{
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		ValidPreparationId:             x.ValidPreparationID,
		ValidIngredientId:              x.ValidIngredientID,
	}
}

func ConvertValidPrepTaskConfigToGRPCValidPrepTaskConfig(x *mealplanning.ValidPrepTaskConfig) *mealplanninggrpc.ValidPrepTaskConfig {
	return &mealplanninggrpc.ValidPrepTaskConfig{
		CreatedAt:                      converters.ConvertTimeToPBTimestamp(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertTimePointerToPBTimestamp(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertTimePointerToPBTimestamp(x.ArchivedAt),
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		Id:                             x.ID,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		Preparation:                    ConvertValidPreparationToGRPCValidPreparation(&x.Preparation),
		Ingredient:                     ConvertValidIngredientToGRPCValidIngredient(&x.Ingredient),
	}
}

func ConvertGRPCValidPrepTaskConfigToValidPrepTaskConfig(x *mealplanninggrpc.ValidPrepTaskConfig) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      converters.ConvertPBTimestampToTime(x.CreatedAt),
		LastUpdatedAt:                  converters.ConvertPBTimestampToTimePointer(x.LastUpdatedAt),
		ArchivedAt:                     converters.ConvertPBTimestampToTimePointer(x.ArchivedAt),
		MinStorageDurationInSeconds:    x.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    x.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		ID:                             x.Id,
		StorageType:                    x.StorageType,
		StorageInstructions:            x.StorageInstructions,
		Notes:                          x.Notes,
		Source:                         x.Source,
		Preparation:                    *ConvertGRPCValidPreparationToValidPreparation(x.Preparation),
		Ingredient:                     *ConvertGRPCValidIngredientToValidIngredient(x.Ingredient),
	}
}
