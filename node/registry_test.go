package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRegistry(t *testing.T) {
	reg := GetRegistry()
	validateRegistryTest(t, reg)
}

func TestNewRegistry(t *testing.T) {
	reg := ConfigRegistry("definition")
	validateRegistryTest(t, reg)
}

func validateRegistryTest(t *testing.T, reg *Registry) {
	assert.Equal(t, len(reg.objs), 288)
	assert.Equal(t, reg.objs[3].Name, "Device")
}

func TestSingleObject(t *testing.T) {
	blob := `
<?xml version="1.0" encoding="UTF-8"?>
<LWM2M  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://openmobilealliance.org/tech/profiles/LWM2M.xsd">
	<Object ObjectType="MODefinition">
		<Name>Temperature</Name>
		<Description1>This IPSO object should be used with a temperature sensor to report a temperature measurement.  It also provides resources for minimum/maximum measured values and the minimum/maximum range that can be measured by the temperature sensor. An example measurement unit is degrees Celsius.</Description1>
		<ObjectID>3303</ObjectID>
		<ObjectURN>urn:oma:github.com/yplam/lwm2m:ext:3303:1.1</ObjectURN>
		<LWM2MVersion>1.0</LWM2MVersion>
		<ObjectVersion>1.1</ObjectVersion>
		<MultipleInstances>Multiple</MultipleInstances>
		<Mandatory>Optional</Mandatory>
		<Resources>
			<Item ID="5700">
				<Name>Sensor Value</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Mandatory</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>Last or Current Measured Value from the Sensor.</Description>
			</Item>
			<Item ID="5601">
				<Name>Min Measured Value</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The minimum value measured by the sensor since power ON or reset.</Description>
			</Item>
			<Item ID="5602">
				<Name>Max Measured Value</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The maximum value measured by the sensor since power ON or reset.</Description>
			</Item>
			<Item ID="5603">
				<Name>Min Range Value</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The minimum value that can be measured by the sensor.</Description>
			</Item>
			<Item ID="5604">
				<Name>Max Range Value</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The maximum value that can be measured by the sensor.</Description>
			</Item>
			<Item ID="5701">
				<Name>Sensor Units</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>String</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>Measurement Units Definition.</Description>
			</Item>
			<Item ID="5605">
				<Name>Reset Min and Max Measured Values</Name>
				<Operations>E</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type></Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>Reset the Min and Max Measured Values to Current Value.</Description>
			</Item>
			<Item ID="5750">
				<Name>Application Type</Name>
				<Operations>RW</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>String</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The application type of the sensor or actuator as a string depending on the use case.</Description>
			</Item>
			<Item ID="5518">
				<Name>Timestamp</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Time</Type>
				<RangeEnumeration></RangeEnumeration>
				<Units></Units>
				<Description>The timestamp of when the measurement was performed.</Description>
			</Item>
			<Item ID="6050">
				<Name>Fractional Timestamp</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Float</Type>
				<RangeEnumeration>0..1</RangeEnumeration>
				<Units>s</Units>
				<Description>Fractional part of the timestamp when sub-second precision is used (e.g., 0.23 for 230 ms).</Description>
			</Item>
			<Item ID="6042">
				<Name>Measurement Quality Indicator</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Integer</Type>
				<RangeEnumeration>0..23</RangeEnumeration>
				<Units></Units>
				<Description>Measurement quality indicator reported by a smart sensor. 0: UNCHECKED No quality checks were done because they do not exist or can not be applied. 1: REJECTED WITH CERTAINTY The measured value is invalid. 2: REJECTED WITH PROBABILITY The measured value is likely invalid. 3: ACCEPTED BUT SUSPICIOUS The measured value is likely OK. 4: ACCEPTED The measured value is OK. 5-15: Reserved for future extensions. 16-23: Vendor specific measurement quality.</Description>
			</Item>
			<Item ID="6049">
				<Name>Measurement Quality Level</Name>
				<Operations>R</Operations>
				<MultipleInstances>Single</MultipleInstances>
				<Mandatory>Optional</Mandatory>
				<Type>Integer</Type>
				<RangeEnumeration>0..100</RangeEnumeration>
				<Units></Units>
				<Description>Measurement quality level reported by a smart sensor. Quality level 100 means that the measurement has fully passed quality check algorithms. Smaller quality levels mean that quality has decreased and the measurement has only partially passed quality check algorithms. The smaller the quality level, the more caution should be used by the application when using the measurement. When the quality level is 0 it means that the measurement should certainly be rejected.</Description>
			</Item>
		</Resources>
		<Description2></Description2>
	</Object>
</LWM2M>
`
	obj, err := loadObjectDefinition([]byte(blob))
	assert.Nil(t, err)
	assert.Equal(t, obj.Id, uint16(3303))
	assert.Equal(t, obj.Name, "Temperature")
	assert.Equal(t, obj.Description, "This IPSO object should be used with a temperature sensor to report a temperature measurement.  It also provides resources for minimum/maximum measured values and the minimum/maximum range that can be measured by the temperature sensor. An example measurement unit is degrees Celsius.")
	assert.Equal(t, obj.Multiple, true)
	assert.Equal(t, obj.Mandatory, false)
	assert.Equal(t, obj.LWM2MVersion, "1.0")
	assert.Equal(t, obj.Version, "1.1")
	assert.Equal(t, len(obj.Resources), 12)
	assert.Equal(t, obj.Resources[5700].ID, uint16(5700))
	assert.Equal(t, obj.Resources[5700].Name, "Sensor Value")
	assert.Equal(t, obj.Resources[5700].Operations, OP_R)
	assert.Equal(t, obj.Resources[5700].Multiple, false)
	assert.Equal(t, obj.Resources[5700].Mandatory, true)
	assert.Equal(t, obj.Resources[5700].Type, R_FLOAT)
	assert.Equal(t, obj.Resources[5700].Description, "Last or Current Measured Value from the Sensor.")
}
