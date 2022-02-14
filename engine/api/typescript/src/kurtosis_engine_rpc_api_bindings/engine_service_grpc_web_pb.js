/**
 * @fileoverview gRPC-Web generated client stub for engine_api
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js')
const proto = {};
proto.engine_api = require('../web/engine_service_pb.d.ts');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.engine_api.EngineServiceClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?grpc.web.ClientOptions} options
 * @constructor
 * @struct
 * @final
 */
proto.engine_api.EngineServicePromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options.format = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.engine_api.GetEngineInfoResponse>}
 */
const methodDescriptor_EngineService_GetEngineInfo = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/GetEngineInfo',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.engine_api.GetEngineInfoResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.engine_api.GetEngineInfoResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.engine_api.GetEngineInfoResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.engine_api.GetEngineInfoResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.getEngineInfo =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/GetEngineInfo',
      request,
      metadata || {},
      methodDescriptor_EngineService_GetEngineInfo,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.engine_api.GetEngineInfoResponse>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.getEngineInfo =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/GetEngineInfo',
      request,
      metadata || {},
      methodDescriptor_EngineService_GetEngineInfo);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.engine_api.CreateEnclaveArgs,
 *   !proto.engine_api.CreateEnclaveResponse>}
 */
const methodDescriptor_EngineService_CreateEnclave = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/CreateEnclave',
  grpc.web.MethodType.UNARY,
  proto.engine_api.CreateEnclaveArgs,
  proto.engine_api.CreateEnclaveResponse,
  /**
   * @param {!proto.engine_api.CreateEnclaveArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.engine_api.CreateEnclaveResponse.deserializeBinary
);


/**
 * @param {!proto.engine_api.CreateEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.engine_api.CreateEnclaveResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.engine_api.CreateEnclaveResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.createEnclave =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/CreateEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_CreateEnclave,
      callback);
};


/**
 * @param {!proto.engine_api.CreateEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.engine_api.CreateEnclaveResponse>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.createEnclave =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/CreateEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_CreateEnclave);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.google.protobuf.Empty,
 *   !proto.engine_api.GetEnclavesResponse>}
 */
const methodDescriptor_EngineService_GetEnclaves = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/GetEnclaves',
  grpc.web.MethodType.UNARY,
  google_protobuf_empty_pb.Empty,
  proto.engine_api.GetEnclavesResponse,
  /**
   * @param {!proto.google.protobuf.Empty} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.engine_api.GetEnclavesResponse.deserializeBinary
);


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.engine_api.GetEnclavesResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.engine_api.GetEnclavesResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.getEnclaves =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/GetEnclaves',
      request,
      metadata || {},
      methodDescriptor_EngineService_GetEnclaves,
      callback);
};


/**
 * @param {!proto.google.protobuf.Empty} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.engine_api.GetEnclavesResponse>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.getEnclaves =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/GetEnclaves',
      request,
      metadata || {},
      methodDescriptor_EngineService_GetEnclaves);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.engine_api.StopEnclaveArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_EngineService_StopEnclave = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/StopEnclave',
  grpc.web.MethodType.UNARY,
  proto.engine_api.StopEnclaveArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.engine_api.StopEnclaveArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.engine_api.StopEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.stopEnclave =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/StopEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_StopEnclave,
      callback);
};


/**
 * @param {!proto.engine_api.StopEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.stopEnclave =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/StopEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_StopEnclave);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.engine_api.DestroyEnclaveArgs,
 *   !proto.google.protobuf.Empty>}
 */
const methodDescriptor_EngineService_DestroyEnclave = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/DestroyEnclave',
  grpc.web.MethodType.UNARY,
  proto.engine_api.DestroyEnclaveArgs,
  google_protobuf_empty_pb.Empty,
  /**
   * @param {!proto.engine_api.DestroyEnclaveArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  google_protobuf_empty_pb.Empty.deserializeBinary
);


/**
 * @param {!proto.engine_api.DestroyEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.google.protobuf.Empty)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.google.protobuf.Empty>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.destroyEnclave =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/DestroyEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_DestroyEnclave,
      callback);
};


/**
 * @param {!proto.engine_api.DestroyEnclaveArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.google.protobuf.Empty>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.destroyEnclave =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/DestroyEnclave',
      request,
      metadata || {},
      methodDescriptor_EngineService_DestroyEnclave);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.engine_api.CleanArgs,
 *   !proto.engine_api.CleanResponse>}
 */
const methodDescriptor_EngineService_Clean = new grpc.web.MethodDescriptor(
  '/engine_api.EngineService/Clean',
  grpc.web.MethodType.UNARY,
  proto.engine_api.CleanArgs,
  proto.engine_api.CleanResponse,
  /**
   * @param {!proto.engine_api.CleanArgs} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.engine_api.CleanResponse.deserializeBinary
);


/**
 * @param {!proto.engine_api.CleanArgs} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.RpcError, ?proto.engine_api.CleanResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.engine_api.CleanResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.engine_api.EngineServiceClient.prototype.clean =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/engine_api.EngineService/Clean',
      request,
      metadata || {},
      methodDescriptor_EngineService_Clean,
      callback);
};


/**
 * @param {!proto.engine_api.CleanArgs} request The
 *     request proto
 * @param {?Object<string, string>=} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.engine_api.CleanResponse>}
 *     Promise that resolves to the response
 */
proto.engine_api.EngineServicePromiseClient.prototype.clean =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/engine_api.EngineService/Clean',
      request,
      metadata || {},
      methodDescriptor_EngineService_Clean);
};


module.exports = proto.engine_api;

