/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"net/http"

	"configcenter/src/common/http/rest"
	"configcenter/src/source_controller/coreservice/service/capability"
	fieldtmpl "configcenter/src/source_controller/coreservice/service/field_template"
	"configcenter/src/source_controller/coreservice/service/id_rule"
	"configcenter/src/source_controller/coreservice/service/kube"
	modelquote "configcenter/src/source_controller/coreservice/service/model_quote"

	"github.com/emicklei/go-restful/v3"
)

func (s *coreService) initModelClassification(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/model/classification",
		Handler: s.CreateOneModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/model/classification",
		Handler: s.CreateManyModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/setmany/model/classification",
		Handler: s.SetManyModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/model/classification",
		Handler: s.SetOneModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model/classification",
		Handler: s.UpdateModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/classification",
		Handler: s.DeleteModelClassification})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model/classification",
		Handler: s.SearchModelClassification})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initModel(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/model", Handler: s.CreateModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path:    "/create/table/model",
		Handler: s.CreateTableModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/model", Handler: s.SetModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model", Handler: s.UpdateModel})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model", Handler: s.DeleteModel})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/{id}/cascade",
		Handler: s.CascadeDeleteModel})
	utility.AddHandler(rest.Action{
		Verb:    http.MethodDelete,
		Path:    "/delete/table/model/cascade",
		Handler: s.CascadeDeleteTableModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model", Handler: s.SearchModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/model", Handler: s.ListModel})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model/with/attribute",
		Handler: s.SearchModelWithAttribute})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/read/model/statistics", Handler: s.GetModelStatistics})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/model/tables", Handler: s.CreateModelTables})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path:    "/create/table/model/tables",
		Handler: s.CreateTableModelTables})

	// init model attribute groups methods
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/create/model/{bk_obj_id}/group", Handler: s.CreateModelAttributeGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/set/model/{bk_obj_id}/group", Handler: s.SetModelAttributeGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPut,
		Path: "/update/model/{bk_obj_id}/group", Handler: s.UpdateModelAttributeGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model/group",
		Handler: s.UpdateModelAttributeGroupByCondition})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/{bk_obj_id}/group",
		Handler: s.DeleteModelAttributeGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete,
		Path: "/delete/model/group", Handler: s.DeleteModelAttributeGroupByCondition})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/model/{bk_obj_id}/group", Handler: s.SearchModelAttributeGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/model/group", Handler: s.SearchModelAttributeGroupByCondition})

	// init attributes methods
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/model/{bk_obj_id}/attributes",
		Handler: s.CreateModelAttributes})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/model/{bk_obj_id}/attributes",
		Handler: s.SetModelAttributes})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model/{bk_obj_id}/attributes",
		Handler: s.UpdateModelAttributes})
	utility.AddHandler(rest.Action{Verb: http.MethodPut,
		Path:    "/update/table/model/attributes",
		Handler: s.UpdateTableModelAttrByCondition})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/model/{bk_obj_id}/attribute/{id}/index",
		Handler: s.UpdateModelAttributesIndex})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model/attributes",
		Handler: s.UpdateModelAttributesByCondition})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/{bk_obj_id}/attributes",
		Handler: s.DeleteModelAttribute})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model/{bk_obj_id}/attributes",
		Handler: s.SearchModelAttributes})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model/attributes",
		Handler: s.SearchModelAttributesByCondition})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path:    "/read/{bk_biz_id}/model/attributes/with_table",
		Handler: s.SearchModelAttrsWithTableByCondition})

	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/idgenerator",
		Handler: s.UpdateIDGenerator})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initAttrUnique(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/model/attributes/unique", Handler: s.SearchModelAttrUnique})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/create/model/{bk_obj_id}/attributes/unique", Handler: s.CreateModelAttrUnique})
	utility.AddHandler(rest.Action{Verb: http.MethodPut,
		Path: "/update/model/{bk_obj_id}/attributes/unique/{id}", Handler: s.UpdateModelAttrUnique})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete,
		Path: "/delete/model/{bk_obj_id}/attributes/unique/{id}", Handler: s.DeleteModelAttrUnique})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initModelInstances(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/model/{bk_obj_id}/instance",
		Handler: s.CreateOneModelInstance})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/model/{bk_obj_id}/instance",
		Handler: s.CreateManyModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/batch/create/model/{bk_obj_id}/instance",
		Handler: s.BatchCreateModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/model/{bk_obj_id}/instance",
		Handler: s.UpdateModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/model/{bk_obj_id}/instances",
		Handler: s.SearchModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/count/model/{bk_obj_id}/instances",
		Handler: s.CountModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/{bk_obj_id}/instance",
		Handler: s.DeleteModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/model/{bk_obj_id}/instance/cascade",
		Handler: s.CascadeDeleteModelInstances})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/get/instance/object/mapping",
		Handler: s.GetInstanceObjectMapping})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initAssociationKind(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/create/associationkind", Handler: s.CreateOneAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/createmany/associationkind", Handler: s.CreateManyAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/set/associationkind", Handler: s.SetOneAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/setmany/associationkind", Handler: s.SetManyAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodPut,
		Path: "/update/associationkind", Handler: s.UpdateAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete,
		Path: "/delete/associationkind", Handler: s.DeleteAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete,
		Path: "/delete/associationkind/cascade", Handler: s.CascadeDeleteAssociationKind})
	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/associationkind", Handler: s.SearchAssociationKind})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initModelAssociation(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/modelassociation",
		Handler: s.CreateModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/mainlinemodelassociation",
		Handler: s.CreateMainlineModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/modelassociation",
		Handler: s.SetModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/modelassociation",
		Handler: s.UpdateModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/modelassociation",
		Handler: s.SearchModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/modelassociation",
		Handler: s.DeleteModelAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/modelassociation/cascade",
		Handler: s.DeleteModelAssociation})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initInstanceAssociation(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/instanceassociation",
		Handler: s.CreateOneInstanceAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/instanceassociation",
		Handler: s.CreateManyInstanceAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/instanceassociation",
		Handler: s.SearchInstanceAssociation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/count/instanceassociation/model/{bk_obj_id}",
		Handler: s.CountInstanceAssociations})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/instanceassociation",
		Handler: s.DeleteInstanceAssociation})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initMainline(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	// add handler for model topo and business topo
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/mainline/model",
		Handler: s.SearchMainlineModelTopo})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/mainline/instance/{bk_biz_id}", Handler: s.SearchMainlineInstanceTopo})

	utility.AddToRestfulWebService(web)
}

// NOCC:golint/fnsize(设计如此)
func (s *coreService) host(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/set/module/host/relation/inner/module", Handler: s.TransferHostToInnerModule})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/module/host/relation/module",
		Handler: s.TransferHostToNormalModule})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/set/module/host/relation/cross/business",
		Handler: s.TransferHostToAnotherBusiness})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/host", Handler: s.DeleteHostFromSystem})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/host/host_module_relations",
		Handler: s.RemoveFromModule})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/read/module/host/relation", Handler: s.GetHostModuleRelation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/host/indentifier", Handler: s.HostIdentifier})

	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/host/{bk_host_id}", Handler: s.GetHostByID})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/hosts/search", Handler: s.GetHosts})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/host/lock", Handler: s.LockHost})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/host/lock", Handler: s.UnlockHost})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/host/lock/search", Handler: s.QueryLockHost})

	// dynamic grouping handlers.
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/dynamicgroup", Handler: s.CreateDynamicGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/dynamicgroup/{bk_biz_id}/{id}",
		Handler: s.UpdateDynamicGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/dynamicgroup/{bk_biz_id}/{id}",
		Handler: s.GetDynamicGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/dynamicgroup/{bk_biz_id}/{id}",
		Handler: s.DeleteDynamicGroup})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/dynamicgroup/search",
		Handler: s.SearchDynamicGroup})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/usercustom/{bk_user}",
		Handler: s.AddUserCustom})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/usercustom/{bk_user}/{id}",
		Handler: s.UpdateUserCustomByID})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/usercustom/user/search/{bk_user}",
		Handler: s.GetUserCustomByUser})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/usercustom/default",
		Handler: s.GetDefaultUserCustom})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/usercustom/default",
		Handler: s.UpdateDefaultUserCustom})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/hosts/favorites/{user}",
		Handler: s.AddHostFavourite})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/hosts/favorites/{user}/{id}",
		Handler: s.UpdateHostFavouriteByID})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/hosts/favorites/{user}/{id}",
		Handler: s.DeleteHostFavouriteByID})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/hosts/favorites/search/{user}",
		Handler: s.ListHostFavourites})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/hosts/favorites/search/{user}/{id}",
		Handler: s.GetHostFavouriteByID})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/meta/hosts/modules/search",
		Handler: s.GetHostModulesIDs})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/hosts/list_hosts",
		Handler: s.ListHosts})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/updatemany/hosts/cloudarea_field",
		Handler: s.UpdateHostCloudAreaField})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/cloudarea/hostcount",
		Handler: s.FindCloudAreaHostCount})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/distinct/host_id/topology/relation",
		Handler: s.GetDistinctHostIDsByTopoRelation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/host/transfer/resource/directory",
		Handler: s.TransferHostResourceDirectory})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) audit(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/auditlog",
		Handler: s.CreateAuditLog})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/read/auditlog",
		Handler: s.SearchAuditLog})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initOperation(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/operation/chart",
		Handler: s.CreateOperationChart})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/operation/chart",
		Handler: s.SearchChartWithPosition})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/operation/chart",
		Handler: s.UpdateOperationChart})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/operation/chart/{id}",
		Handler: s.DeleteOperationChart})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/operation/chart/common",
		Handler: s.SearchChartCommon})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/operation/inst/count",
		Handler: s.SearchInstCount})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/operation/chart/data",
		Handler: s.SearchChartData})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/operation/chart/position",
		Handler: s.UpdateChartPosition})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/operation/timer/chart/data",
		Handler: s.SearchTimerChartData})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/start/operation/chart/timer",
		Handler: s.TimerFreshData})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) label(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/labels", Handler: s.AddLabels})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/deletemany/labels", Handler: s.RemoveLabels})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/updatemany/labels", Handler: s.UpdateLabels})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) topographics(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/topographics/search", Handler: s.SearchTopoGraphics})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/topographics/update", Handler: s.UpdateTopoGraphics})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) ccSystem(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/system/user_config",
		Handler: s.GetSystemUserConfig})

	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/system/config_admin",
		Handler: s.SearchConfigAdmin})
	utility.AddHandler(rest.Action{Verb: http.MethodGet,
		Path:    "/find/system_config/platform_setting",
		Handler: s.SearchPlatformSettingConfig})

	utility.AddHandler(rest.Action{Verb: http.MethodPut,
		Path:    "/update/system_config/platform_setting",
		Handler: s.UpdatePlatformSetting})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) transaction(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/transaction/commit",
		Handler: s.CommitTransaction})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/transaction/abort",
		Handler: s.AbortTransaction})
	utility.AddToRestfulWebService(web)
}

func (s *coreService) initCount(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/resource/count",
		Handler: s.GetCountByFilter})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initCloudSync(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/cloud/account", Handler: s.CreateAccount})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/cloud/account", Handler: s.SearchAccount})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/cloud/account/{bk_account_id}",
		Handler: s.UpdateAccount})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/cloud/account/{bk_account_id}",
		Handler: s.DeleteAccount})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/cloud/accountconf",
		Handler: s.SearchAccountConf})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/cloud/sync/task", Handler: s.CreateSyncTask})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/cloud/sync/task", Handler: s.SearchSyncTask})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/cloud/sync/task/{bk_task_id}",
		Handler: s.UpdateSyncTask})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/cloud/sync/task/{bk_task_id}",
		Handler: s.DeleteSyncTask})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/cloud/sync/history",
		Handler: s.CreateSyncHistory})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/cloud/sync/history",
		Handler: s.SearchSyncHistory})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/cloud/sync/destroyed_host_related",
		Handler: s.DeleteDestroyedHostRelated})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initAuth(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/search/auth/resource",
		Handler: s.SearchAuthResource})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initCommon(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPost,
		Path: "/findmany/common/distinct_field", Handler: s.GetDistinctField})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/common/distinct_count",
		Handler: s.GetDistinctCount})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/group/related/{kind}/resource/by_ids",
		Handler: s.GroupRelResByIDs})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initProject(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/project/bk_project_id",
		Handler: s.UpdateProjectID})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initModelQuote(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	// model quote relation
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/list/model/quote/relation",
		Handler: s.ListModelQuoteRelation})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/model/quote/relation",
		Handler: s.CreateModelQuoteRelation})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/deletemany/model/quote/relation",
		Handler: s.DeleteModelQuoteRelation})

	// quoted instance
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/createmany/quoted/model/{bk_obj_id}/instance",
		Handler: modelquote.BatchCreateQuotedInstance})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/findmany/quoted/model/{bk_obj_id}/instance",
		Handler: modelquote.ListQuotedInstance})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/updatemany/quoted/model/{bk_obj_id}/instance",
		Handler: modelquote.BatchUpdateQuotedInstance})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/deletemany/quoted/model/{bk_obj_id}/instance",
		Handler: modelquote.BatchDeleteQuotedInstance})

	utility.AddToRestfulWebService(web)
}

func (s *coreService) initService(web *restful.WebService) {
	c := &capability.Capability{
		Utility: rest.NewRestUtility(rest.Config{ErrorIf: s.engine.CCErr, Language: s.engine.Language}),
		Core:    s.core,
	}

	s.initModelClassification(web)
	s.initModel(web)
	s.initAssociationKind(web)
	s.initAttrUnique(web)
	s.initModelAssociation(web)
	s.initModelInstances(web)
	s.initInstanceAssociation(web)
	s.initDataSynchronize(web)
	s.initMainline(web)
	s.host(web)
	s.audit(web)
	s.initProcess(web)
	s.initOperation(web)
	s.label(web)
	s.topographics(web)
	s.ccSystem(web)
	s.initSetTemplate(web)
	s.initHostApplyRule(web)
	s.transaction(web)
	s.initCount(web)
	s.initCloudSync(web)
	s.initAuth(web)
	s.initCommon(web)
	s.initUserManagement(web)
	kube.InitKube(c)
	s.initProject(web)
	s.initModelQuote(web)
	fieldtmpl.InitFieldTemplate(c)
	idrule.InitIDRule(c)

	c.Utility.AddToRestfulWebService(web)
}

func (s *coreService) initUserManagement(web *restful.WebService) {
	utility := rest.NewRestUtility(rest.Config{
		ErrorIf:  s.engine.CCErr,
		Language: s.engine.Language,
	})

	// 用户管理API
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/user", Handler: s.CreateUser})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/user/{user_id}", Handler: s.UpdateUser})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/user/{user_id}", Handler: s.DeleteUser})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/user/{user_id}", Handler: s.GetUser})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/find/users", Handler: s.ListUsers})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/users/batch", Handler: s.BatchDeleteUsers})
	
	// 用户状态管理API
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/user/{user_id}/status", Handler: s.ToggleUserStatus})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/update/user/{user_id}/reset-password", Handler: s.ResetUserPassword})
	
	// 用户统计和查询API
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/users/statistics", Handler: s.GetUserStatistics})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/validate/email", Handler: s.ValidateEmail})
	
	// 用户导入导出API
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/export/users", Handler: s.ExportUsers})
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/import/users", Handler: s.ImportUsers})
	
	// 角色权限管理API
	utility.AddHandler(rest.Action{Verb: http.MethodPost, Path: "/create/role-permission", Handler: s.CreateRolePermission})
	utility.AddHandler(rest.Action{Verb: http.MethodPut, Path: "/update/role-permission/{role_id}", Handler: s.UpdateRolePermission})
	utility.AddHandler(rest.Action{Verb: http.MethodDelete, Path: "/delete/role-permission/{role_id}", Handler: s.DeleteRolePermission})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/role-permission/{role_id}", Handler: s.GetRolePermission})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/role-permissions", Handler: s.ListRolePermissions})
	
	// 权限矩阵API
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/permission-matrix", Handler: s.GetPermissionMatrix})
	utility.AddHandler(rest.Action{Verb: http.MethodGet, Path: "/find/role/{role_id}/users", Handler: s.GetUserRoles})

	utility.AddToRestfulWebService(web)
}
