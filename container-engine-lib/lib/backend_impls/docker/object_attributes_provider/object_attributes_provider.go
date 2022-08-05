package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	engineServerNamePrefix        = "kurtosis-engine"
	logsDatabaseServerNamePrefix  = "kurtosis-logs-db"
	logsCollectorServerNamePrefix = "kurtosis-logs-collector"
)

type DockerObjectAttributesProvider interface {
	ForEngineServer(
		guid engine.EngineGUID,
		grpcPortId string,
		grpcPortSpec *port_spec.PortSpec,
		grpcProxyPortId string,
		grpcProxyPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForEnclave(enclaveId enclave.EnclaveID) (DockerEnclaveObjectAttributesProvider, error)
	ForLogsDatabaseServer(
		guid engine.EngineGUID,
		httpApiPortId string,
		httpApiPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForLogsCollectorServer(
		guid engine.EngineGUID,
		forwardPortId string,
		forwardPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
}

func GetDockerObjectAttributesProvider() DockerObjectAttributesProvider {
	return newDockerObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type dockerObjectAttributesProviderImpl struct{}

func newDockerObjectAttributesProviderImpl() *dockerObjectAttributesProviderImpl {
	return &dockerObjectAttributesProviderImpl{}
}

func (provider *dockerObjectAttributesProviderImpl) ForEngineServer(
	guid engine.EngineGUID,
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {

	nameStr := strings.Join(
		[]string{
			engineServerNamePrefix,
			string(guid),
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		grpcPortId:      grpcPortSpec,
		grpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.ContainerTypeDockerLabelKey: label_value_consts.EngineContainerTypeDockerLabelValue,
		label_key_consts.PortSpecsDockerLabelKey:     serializedPortsSpec,
		label_key_consts.IDDockerLabelKey:            idLabelValue,
		label_key_consts.GUIDDockerLabelKey:          guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForEnclave(enclaveId enclave.EnclaveID) (DockerEnclaveObjectAttributesProvider, error) {
	enclaveIdStr := string(enclaveId)
	enclaveIdLabelValue, err := docker_label_value.CreateNewDockerLabelValue(enclaveIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value out of enclave ID string '%v'", enclaveIdStr)
	}

	return newDockerEnclaveObjectAttributesProviderImpl(enclaveIdLabelValue), nil
}

func (provider *dockerObjectAttributesProviderImpl) ForLogsDatabaseServer(
	guid engine.EngineGUID,
	httpApiPortId string,
	httpApiPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	nameStr := strings.Join(
		[]string{
			logsDatabaseServerNamePrefix,
			string(guid),
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		httpApiPortId: httpApiPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following logs-database-server-ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.ContainerTypeDockerLabelKey: label_value_consts.LogsDatabaseTypeDockerLabelValue,
		label_key_consts.PortSpecsDockerLabelKey:     serializedPortsSpec,
		label_key_consts.IDDockerLabelKey:            idLabelValue,
		label_key_consts.GUIDDockerLabelKey:          guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForLogsCollectorServer(
	guid engine.EngineGUID,
	forwardPortId string,
	forwardPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	nameStr := strings.Join(
		[]string{
			logsCollectorServerNamePrefix,
			string(guid),
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		forwardPortId: forwardPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following logs-collector-server-ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.ContainerTypeDockerLabelKey: label_value_consts.LogsCollectorTypeDockerLabelValue,
		label_key_consts.PortSpecsDockerLabelKey:     serializedPortsSpec,
		label_key_consts.IDDockerLabelKey:            idLabelValue,
		label_key_consts.GUIDDockerLabelKey:          guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}
