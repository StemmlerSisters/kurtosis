// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.8
// source: api_container_service.proto

package kurtosis_core_rpc_api_bindings

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ApiContainerServiceClient is the client API for ApiContainerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApiContainerServiceClient interface {
	// Starts a module container in the enclave
	LoadModule(ctx context.Context, in *LoadModuleArgs, opts ...grpc.CallOption) (*LoadModuleResponse, error)
	// Gets information about loaded modules
	GetModules(ctx context.Context, in *GetModulesArgs, opts ...grpc.CallOption) (*GetModulesResponse, error)
	// Stop and remove a module from the enclave
	UnloadModule(ctx context.Context, in *UnloadModuleArgs, opts ...grpc.CallOption) (*UnloadModuleResponse, error)
	// Executes an executable module on the user's behalf
	ExecuteModule(ctx context.Context, in *ExecuteModuleArgs, opts ...grpc.CallOption) (*ExecuteModuleResponse, error)
	// Executes a startosis script on the user's behalf
	ExecuteStartosisScript(ctx context.Context, in *ExecuteStartosisScriptArgs, opts ...grpc.CallOption) (*ExecuteStartosisResponse, error)
	// Executes a startosis module on the user's behalf
	ExecuteStartosisModule(ctx context.Context, in *ExecuteStartosisModuleArgs, opts ...grpc.CallOption) (*ExecuteStartosisResponse, error)
	// Start services by creating containers for them
	StartServices(ctx context.Context, in *StartServicesArgs, opts ...grpc.CallOption) (*StartServicesResponse, error)
	// Returns the IDs of the current services in the enclave
	GetServices(ctx context.Context, in *GetServicesArgs, opts ...grpc.CallOption) (*GetServicesResponse, error)
	// Instructs the API container to remove the given service
	RemoveService(ctx context.Context, in *RemoveServiceArgs, opts ...grpc.CallOption) (*RemoveServiceResponse, error)
	// Instructs the API container to repartition the enclave
	Repartition(ctx context.Context, in *RepartitionArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Executes the given command inside a running container
	ExecCommand(ctx context.Context, in *ExecCommandArgs, opts ...grpc.CallOption) (*ExecCommandResponse, error)
	// Pauses all processes running in the service container
	PauseService(ctx context.Context, in *PauseServiceArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Unpauses all paused processes running in the service container
	UnpauseService(ctx context.Context, in *UnpauseServiceArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
	WaitForHttpGetEndpointAvailability(ctx context.Context, in *WaitForHttpGetEndpointAvailabilityArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
	WaitForHttpPostEndpointAvailability(ctx context.Context, in *WaitForHttpPostEndpointAvailabilityArgs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Uploads a files artifact to the Kurtosis File System
	UploadFilesArtifact(ctx context.Context, in *UploadFilesArtifactArgs, opts ...grpc.CallOption) (*UploadFilesArtifactResponse, error)
	// TODO Make this a server-side streaming method so the client can download large files
	// Downloads a files artifact from the Kurtosis File System
	DownloadFilesArtifact(ctx context.Context, in *DownloadFilesArtifactArgs, opts ...grpc.CallOption) (*DownloadFilesArtifactResponse, error)
	// Tells the API container to download a files artifact from the web to the Kurtosis File System
	StoreWebFilesArtifact(ctx context.Context, in *StoreWebFilesArtifactArgs, opts ...grpc.CallOption) (*StoreWebFilesArtifactResponse, error)
	// Tells the API container to copy a files artifact from a service to the Kurtosis File System
	StoreFilesArtifactFromService(ctx context.Context, in *StoreFilesArtifactFromServiceArgs, opts ...grpc.CallOption) (*StoreFilesArtifactFromServiceResponse, error)
	// Renders the templates and their data to a files artifact in the Kurtosis File System
	RenderTemplatesToFilesArtifact(ctx context.Context, in *RenderTemplatesToFilesArtifactArgs, opts ...grpc.CallOption) (*RenderTemplatesToFilesArtifactResponse, error)
}

type apiContainerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewApiContainerServiceClient(cc grpc.ClientConnInterface) ApiContainerServiceClient {
	return &apiContainerServiceClient{cc}
}

func (c *apiContainerServiceClient) LoadModule(ctx context.Context, in *LoadModuleArgs, opts ...grpc.CallOption) (*LoadModuleResponse, error) {
	out := new(LoadModuleResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/LoadModule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) GetModules(ctx context.Context, in *GetModulesArgs, opts ...grpc.CallOption) (*GetModulesResponse, error) {
	out := new(GetModulesResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/GetModules", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) UnloadModule(ctx context.Context, in *UnloadModuleArgs, opts ...grpc.CallOption) (*UnloadModuleResponse, error) {
	out := new(UnloadModuleResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/UnloadModule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) ExecuteModule(ctx context.Context, in *ExecuteModuleArgs, opts ...grpc.CallOption) (*ExecuteModuleResponse, error) {
	out := new(ExecuteModuleResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/ExecuteModule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) ExecuteStartosisScript(ctx context.Context, in *ExecuteStartosisScriptArgs, opts ...grpc.CallOption) (*ExecuteStartosisResponse, error) {
	out := new(ExecuteStartosisResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/ExecuteStartosisScript", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) ExecuteStartosisModule(ctx context.Context, in *ExecuteStartosisModuleArgs, opts ...grpc.CallOption) (*ExecuteStartosisResponse, error) {
	out := new(ExecuteStartosisResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/ExecuteStartosisModule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) StartServices(ctx context.Context, in *StartServicesArgs, opts ...grpc.CallOption) (*StartServicesResponse, error) {
	out := new(StartServicesResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/StartServices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) GetServices(ctx context.Context, in *GetServicesArgs, opts ...grpc.CallOption) (*GetServicesResponse, error) {
	out := new(GetServicesResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/GetServices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) RemoveService(ctx context.Context, in *RemoveServiceArgs, opts ...grpc.CallOption) (*RemoveServiceResponse, error) {
	out := new(RemoveServiceResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/RemoveService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) Repartition(ctx context.Context, in *RepartitionArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/Repartition", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) ExecCommand(ctx context.Context, in *ExecCommandArgs, opts ...grpc.CallOption) (*ExecCommandResponse, error) {
	out := new(ExecCommandResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/ExecCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) PauseService(ctx context.Context, in *PauseServiceArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/PauseService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) UnpauseService(ctx context.Context, in *UnpauseServiceArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/UnpauseService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) WaitForHttpGetEndpointAvailability(ctx context.Context, in *WaitForHttpGetEndpointAvailabilityArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) WaitForHttpPostEndpointAvailability(ctx context.Context, in *WaitForHttpPostEndpointAvailabilityArgs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) UploadFilesArtifact(ctx context.Context, in *UploadFilesArtifactArgs, opts ...grpc.CallOption) (*UploadFilesArtifactResponse, error) {
	out := new(UploadFilesArtifactResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/UploadFilesArtifact", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) DownloadFilesArtifact(ctx context.Context, in *DownloadFilesArtifactArgs, opts ...grpc.CallOption) (*DownloadFilesArtifactResponse, error) {
	out := new(DownloadFilesArtifactResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/DownloadFilesArtifact", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) StoreWebFilesArtifact(ctx context.Context, in *StoreWebFilesArtifactArgs, opts ...grpc.CallOption) (*StoreWebFilesArtifactResponse, error) {
	out := new(StoreWebFilesArtifactResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/StoreWebFilesArtifact", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) StoreFilesArtifactFromService(ctx context.Context, in *StoreFilesArtifactFromServiceArgs, opts ...grpc.CallOption) (*StoreFilesArtifactFromServiceResponse, error) {
	out := new(StoreFilesArtifactFromServiceResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/StoreFilesArtifactFromService", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiContainerServiceClient) RenderTemplatesToFilesArtifact(ctx context.Context, in *RenderTemplatesToFilesArtifactArgs, opts ...grpc.CallOption) (*RenderTemplatesToFilesArtifactResponse, error) {
	out := new(RenderTemplatesToFilesArtifactResponse)
	err := c.cc.Invoke(ctx, "/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApiContainerServiceServer is the server API for ApiContainerService service.
// All implementations should embed UnimplementedApiContainerServiceServer
// for forward compatibility
type ApiContainerServiceServer interface {
	// Starts a module container in the enclave
	LoadModule(context.Context, *LoadModuleArgs) (*LoadModuleResponse, error)
	// Gets information about loaded modules
	GetModules(context.Context, *GetModulesArgs) (*GetModulesResponse, error)
	// Stop and remove a module from the enclave
	UnloadModule(context.Context, *UnloadModuleArgs) (*UnloadModuleResponse, error)
	// Executes an executable module on the user's behalf
	ExecuteModule(context.Context, *ExecuteModuleArgs) (*ExecuteModuleResponse, error)
	// Executes a startosis script on the user's behalf
	ExecuteStartosisScript(context.Context, *ExecuteStartosisScriptArgs) (*ExecuteStartosisResponse, error)
	// Executes a startosis module on the user's behalf
	ExecuteStartosisModule(context.Context, *ExecuteStartosisModuleArgs) (*ExecuteStartosisResponse, error)
	// Start services by creating containers for them
	StartServices(context.Context, *StartServicesArgs) (*StartServicesResponse, error)
	// Returns the IDs of the current services in the enclave
	GetServices(context.Context, *GetServicesArgs) (*GetServicesResponse, error)
	// Instructs the API container to remove the given service
	RemoveService(context.Context, *RemoveServiceArgs) (*RemoveServiceResponse, error)
	// Instructs the API container to repartition the enclave
	Repartition(context.Context, *RepartitionArgs) (*emptypb.Empty, error)
	// Executes the given command inside a running container
	ExecCommand(context.Context, *ExecCommandArgs) (*ExecCommandResponse, error)
	// Pauses all processes running in the service container
	PauseService(context.Context, *PauseServiceArgs) (*emptypb.Empty, error)
	// Unpauses all paused processes running in the service container
	UnpauseService(context.Context, *UnpauseServiceArgs) (*emptypb.Empty, error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
	WaitForHttpGetEndpointAvailability(context.Context, *WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
	WaitForHttpPostEndpointAvailability(context.Context, *WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error)
	// Uploads a files artifact to the Kurtosis File System
	UploadFilesArtifact(context.Context, *UploadFilesArtifactArgs) (*UploadFilesArtifactResponse, error)
	// TODO Make this a server-side streaming method so the client can download large files
	// Downloads a files artifact from the Kurtosis File System
	DownloadFilesArtifact(context.Context, *DownloadFilesArtifactArgs) (*DownloadFilesArtifactResponse, error)
	// Tells the API container to download a files artifact from the web to the Kurtosis File System
	StoreWebFilesArtifact(context.Context, *StoreWebFilesArtifactArgs) (*StoreWebFilesArtifactResponse, error)
	// Tells the API container to copy a files artifact from a service to the Kurtosis File System
	StoreFilesArtifactFromService(context.Context, *StoreFilesArtifactFromServiceArgs) (*StoreFilesArtifactFromServiceResponse, error)
	// Renders the templates and their data to a files artifact in the Kurtosis File System
	RenderTemplatesToFilesArtifact(context.Context, *RenderTemplatesToFilesArtifactArgs) (*RenderTemplatesToFilesArtifactResponse, error)
}

// UnimplementedApiContainerServiceServer should be embedded to have forward compatible implementations.
type UnimplementedApiContainerServiceServer struct {
}

func (UnimplementedApiContainerServiceServer) LoadModule(context.Context, *LoadModuleArgs) (*LoadModuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadModule not implemented")
}
func (UnimplementedApiContainerServiceServer) GetModules(context.Context, *GetModulesArgs) (*GetModulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModules not implemented")
}
func (UnimplementedApiContainerServiceServer) UnloadModule(context.Context, *UnloadModuleArgs) (*UnloadModuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnloadModule not implemented")
}
func (UnimplementedApiContainerServiceServer) ExecuteModule(context.Context, *ExecuteModuleArgs) (*ExecuteModuleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteModule not implemented")
}
func (UnimplementedApiContainerServiceServer) ExecuteStartosisScript(context.Context, *ExecuteStartosisScriptArgs) (*ExecuteStartosisResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteStartosisScript not implemented")
}
func (UnimplementedApiContainerServiceServer) ExecuteStartosisModule(context.Context, *ExecuteStartosisModuleArgs) (*ExecuteStartosisResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteStartosisModule not implemented")
}
func (UnimplementedApiContainerServiceServer) StartServices(context.Context, *StartServicesArgs) (*StartServicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartServices not implemented")
}
func (UnimplementedApiContainerServiceServer) GetServices(context.Context, *GetServicesArgs) (*GetServicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServices not implemented")
}
func (UnimplementedApiContainerServiceServer) RemoveService(context.Context, *RemoveServiceArgs) (*RemoveServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveService not implemented")
}
func (UnimplementedApiContainerServiceServer) Repartition(context.Context, *RepartitionArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Repartition not implemented")
}
func (UnimplementedApiContainerServiceServer) ExecCommand(context.Context, *ExecCommandArgs) (*ExecCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecCommand not implemented")
}
func (UnimplementedApiContainerServiceServer) PauseService(context.Context, *PauseServiceArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseService not implemented")
}
func (UnimplementedApiContainerServiceServer) UnpauseService(context.Context, *UnpauseServiceArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnpauseService not implemented")
}
func (UnimplementedApiContainerServiceServer) WaitForHttpGetEndpointAvailability(context.Context, *WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitForHttpGetEndpointAvailability not implemented")
}
func (UnimplementedApiContainerServiceServer) WaitForHttpPostEndpointAvailability(context.Context, *WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitForHttpPostEndpointAvailability not implemented")
}
func (UnimplementedApiContainerServiceServer) UploadFilesArtifact(context.Context, *UploadFilesArtifactArgs) (*UploadFilesArtifactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UploadFilesArtifact not implemented")
}
func (UnimplementedApiContainerServiceServer) DownloadFilesArtifact(context.Context, *DownloadFilesArtifactArgs) (*DownloadFilesArtifactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadFilesArtifact not implemented")
}
func (UnimplementedApiContainerServiceServer) StoreWebFilesArtifact(context.Context, *StoreWebFilesArtifactArgs) (*StoreWebFilesArtifactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StoreWebFilesArtifact not implemented")
}
func (UnimplementedApiContainerServiceServer) StoreFilesArtifactFromService(context.Context, *StoreFilesArtifactFromServiceArgs) (*StoreFilesArtifactFromServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StoreFilesArtifactFromService not implemented")
}
func (UnimplementedApiContainerServiceServer) RenderTemplatesToFilesArtifact(context.Context, *RenderTemplatesToFilesArtifactArgs) (*RenderTemplatesToFilesArtifactResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RenderTemplatesToFilesArtifact not implemented")
}

// UnsafeApiContainerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApiContainerServiceServer will
// result in compilation errors.
type UnsafeApiContainerServiceServer interface {
	mustEmbedUnimplementedApiContainerServiceServer()
}

func RegisterApiContainerServiceServer(s grpc.ServiceRegistrar, srv ApiContainerServiceServer) {
	s.RegisterService(&ApiContainerService_ServiceDesc, srv)
}

func _ApiContainerService_LoadModule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadModuleArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).LoadModule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/LoadModule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).LoadModule(ctx, req.(*LoadModuleArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_GetModules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetModulesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).GetModules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/GetModules",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).GetModules(ctx, req.(*GetModulesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_UnloadModule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnloadModuleArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).UnloadModule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/UnloadModule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).UnloadModule(ctx, req.(*UnloadModuleArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_ExecuteModule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteModuleArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).ExecuteModule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/ExecuteModule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).ExecuteModule(ctx, req.(*ExecuteModuleArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_ExecuteStartosisScript_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteStartosisScriptArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).ExecuteStartosisScript(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/ExecuteStartosisScript",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).ExecuteStartosisScript(ctx, req.(*ExecuteStartosisScriptArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_ExecuteStartosisModule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteStartosisModuleArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).ExecuteStartosisModule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/ExecuteStartosisModule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).ExecuteStartosisModule(ctx, req.(*ExecuteStartosisModuleArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_StartServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartServicesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).StartServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/StartServices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).StartServices(ctx, req.(*StartServicesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_GetServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetServicesArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).GetServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/GetServices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).GetServices(ctx, req.(*GetServicesArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_RemoveService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveServiceArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).RemoveService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/RemoveService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).RemoveService(ctx, req.(*RemoveServiceArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_Repartition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepartitionArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).Repartition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/Repartition",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).Repartition(ctx, req.(*RepartitionArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_ExecCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecCommandArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).ExecCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/ExecCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).ExecCommand(ctx, req.(*ExecCommandArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_PauseService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PauseServiceArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).PauseService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/PauseService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).PauseService(ctx, req.(*PauseServiceArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_UnpauseService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnpauseServiceArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).UnpauseService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/UnpauseService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).UnpauseService(ctx, req.(*UnpauseServiceArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_WaitForHttpGetEndpointAvailability_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitForHttpGetEndpointAvailabilityArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).WaitForHttpGetEndpointAvailability(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).WaitForHttpGetEndpointAvailability(ctx, req.(*WaitForHttpGetEndpointAvailabilityArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_WaitForHttpPostEndpointAvailability_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitForHttpPostEndpointAvailabilityArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).WaitForHttpPostEndpointAvailability(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).WaitForHttpPostEndpointAvailability(ctx, req.(*WaitForHttpPostEndpointAvailabilityArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_UploadFilesArtifact_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadFilesArtifactArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).UploadFilesArtifact(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/UploadFilesArtifact",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).UploadFilesArtifact(ctx, req.(*UploadFilesArtifactArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_DownloadFilesArtifact_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadFilesArtifactArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).DownloadFilesArtifact(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/DownloadFilesArtifact",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).DownloadFilesArtifact(ctx, req.(*DownloadFilesArtifactArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_StoreWebFilesArtifact_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StoreWebFilesArtifactArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).StoreWebFilesArtifact(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/StoreWebFilesArtifact",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).StoreWebFilesArtifact(ctx, req.(*StoreWebFilesArtifactArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_StoreFilesArtifactFromService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StoreFilesArtifactFromServiceArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).StoreFilesArtifactFromService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/StoreFilesArtifactFromService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).StoreFilesArtifactFromService(ctx, req.(*StoreFilesArtifactFromServiceArgs))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApiContainerService_RenderTemplatesToFilesArtifact_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenderTemplatesToFilesArtifactArgs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApiContainerServiceServer).RenderTemplatesToFilesArtifact(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api_container_api.ApiContainerService/RenderTemplatesToFilesArtifact",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApiContainerServiceServer).RenderTemplatesToFilesArtifact(ctx, req.(*RenderTemplatesToFilesArtifactArgs))
	}
	return interceptor(ctx, in, info, handler)
}

// ApiContainerService_ServiceDesc is the grpc.ServiceDesc for ApiContainerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ApiContainerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api_container_api.ApiContainerService",
	HandlerType: (*ApiContainerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LoadModule",
			Handler:    _ApiContainerService_LoadModule_Handler,
		},
		{
			MethodName: "GetModules",
			Handler:    _ApiContainerService_GetModules_Handler,
		},
		{
			MethodName: "UnloadModule",
			Handler:    _ApiContainerService_UnloadModule_Handler,
		},
		{
			MethodName: "ExecuteModule",
			Handler:    _ApiContainerService_ExecuteModule_Handler,
		},
		{
			MethodName: "ExecuteStartosisScript",
			Handler:    _ApiContainerService_ExecuteStartosisScript_Handler,
		},
		{
			MethodName: "ExecuteStartosisModule",
			Handler:    _ApiContainerService_ExecuteStartosisModule_Handler,
		},
		{
			MethodName: "StartServices",
			Handler:    _ApiContainerService_StartServices_Handler,
		},
		{
			MethodName: "GetServices",
			Handler:    _ApiContainerService_GetServices_Handler,
		},
		{
			MethodName: "RemoveService",
			Handler:    _ApiContainerService_RemoveService_Handler,
		},
		{
			MethodName: "Repartition",
			Handler:    _ApiContainerService_Repartition_Handler,
		},
		{
			MethodName: "ExecCommand",
			Handler:    _ApiContainerService_ExecCommand_Handler,
		},
		{
			MethodName: "PauseService",
			Handler:    _ApiContainerService_PauseService_Handler,
		},
		{
			MethodName: "UnpauseService",
			Handler:    _ApiContainerService_UnpauseService_Handler,
		},
		{
			MethodName: "WaitForHttpGetEndpointAvailability",
			Handler:    _ApiContainerService_WaitForHttpGetEndpointAvailability_Handler,
		},
		{
			MethodName: "WaitForHttpPostEndpointAvailability",
			Handler:    _ApiContainerService_WaitForHttpPostEndpointAvailability_Handler,
		},
		{
			MethodName: "UploadFilesArtifact",
			Handler:    _ApiContainerService_UploadFilesArtifact_Handler,
		},
		{
			MethodName: "DownloadFilesArtifact",
			Handler:    _ApiContainerService_DownloadFilesArtifact_Handler,
		},
		{
			MethodName: "StoreWebFilesArtifact",
			Handler:    _ApiContainerService_StoreWebFilesArtifact_Handler,
		},
		{
			MethodName: "StoreFilesArtifactFromService",
			Handler:    _ApiContainerService_StoreFilesArtifactFromService_Handler,
		},
		{
			MethodName: "RenderTemplatesToFilesArtifact",
			Handler:    _ApiContainerService_RenderTemplatesToFilesArtifact_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api_container_service.proto",
}
