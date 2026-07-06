// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/base64"
	"encoding/xml"
)

// AuthnRequest represents a SAML authentication request with relevant SAML protocol attributes and elements.
type AuthnRequest struct {
	XMLName                        xml.Name
	SAMLP                          string                `xml:"xmlns:samlp,attr"`
	SAML                           string                `xml:"xmlns:saml,attr"`
	SAMLSIG                        string                `xml:"xmlns:samlsig,attr"`
	ID                             string                `xml:"ID,attr"`
	Destination                    string                `xml:"Destination,attr"`
	Version                        string                `xml:"Version,attr"`
	ProtocolBinding                string                `xml:"ProtocolBinding,attr"`
	AssertionConsumerServiceURL    string                `xml:"AssertionConsumerServiceURL,attr"`
	IssueInstant                   string                `xml:"IssueInstant,attr"`
	AssertionConsumerServiceIndex  int                   `xml:"AssertionConsumerServiceIndex,attr"`
	AttributeConsumingServiceIndex int                   `xml:"AttributeConsumingServiceIndex,attr"`
	Issuer                         Issuer                `xml:"Issuer"`
	NameIDPolicy                   NameIDPolicy          `xml:"NameIDPolicy"`
	RequestedAuthnContext          RequestedAuthnContext `xml:"RequestedAuthnContext"`
	Signature                      Signature             `xml:"Signature,omitempty"`
	originalString                 string
}

// Issuer represents the SAML issuer element, typically containing the issuing entity's identifier.
type Issuer struct {
	XMLName xml.Name
	SAML    string `xml:"xmlns:saml2,attr"`
	URL     string `xml:",innerxml"`
}

// NameIDPolicy represents the SAML NameIDPolicy element, defining policy for name identifier format.
type NameIDPolicy struct {
	XMLName     xml.Name
	AllowCreate bool   `xml:"AllowCreate,attr"`
	Format      string `xml:"Format,attr"`
}

// RequestedAuthnContext represents the SAML RequestedAuthnContext element, which specifies authentication requireme
type RequestedAuthnContext struct {
	XMLName              xml.Name
	SAMLP                string               `xml:"xmlns:samlp,attr"`
	Comparison           string               `xml:"Comparison,attr"`
	AuthnContextClassRef AuthnContextClassRef `xml:"AuthnContextClassRef"`
}

// AuthnContextClassRef represents the SAML AuthnContextClassRef element, defining a specific context class.
type AuthnContextClassRef struct {
	XMLName   xml.Name
	SAML      string `xml:"xmlns:saml,attr"`
	Transport string `xml:",innerxml"`
}

// Signature represents a digital signature used in SAML responses and assertions.
type Signature struct {
	XMLName        xml.Name
	ID             string `xml:"Id,attr"`
	SignedInfo     SignedInfo
	SignatureValue SignatureValue
	KeyInfo        KeyInfo
	DS             string `xml:"xmlns:dsig,attr"`
}

// SignedInfo contains the information about the signature method and references in a digital signature.
type SignedInfo struct {
	XMLName                xml.Name
	CanonicalizationMethod CanonicalizationMethod
	SignatureMethod        SignatureMethod
	SamlsigReference       SamlsigReference
}

// SignatureValue represents the actual digital signature value in a SAML message.
type SignatureValue struct {
	XMLName xml.Name
	Value   string `xml:",innerxml"`
}

// KeyInfo contains key information, typically a public key, associated with a digital signature.
type KeyInfo struct {
	XMLName  xml.Name
	X509Data X509Data
}

// KeyInfoMain provides main key information used in encrypted assertions, including an encrypted key.
type KeyInfoMain struct {
	XMLName      xml.Name     `xml:"KeyInfo"`
	EncryptedKey EncryptedKey `xml:"EncryptedKey,omitempty"`
}

// CanonicalizationMethod represents the method used to canonicalize the XML before signing.
type CanonicalizationMethod struct {
	XMLName   xml.Name
	Algorithm string `xml:"Algorithm,attr"`
}

// SignatureMethod represents the algorithm used for generating the digital signature.
type SignatureMethod struct {
	XMLName   xml.Name
	Algorithm string `xml:"Algorithm,attr"`
}

// SamlsigReference represents a reference element in the signature, including transformations and digest information.
type SamlsigReference struct {
	XMLName      xml.Name
	URI          string       `xml:"URI,attr"`
	Transforms   Transforms   `xml:",innerxml"`
	DigestMethod DigestMethod `xml:",innerxml"`
	DigestValue  DigestValue  `xml:",innerxml"`
}

// X509Data represents X.509 certificate information used for validating a signature.
type X509Data struct {
	XMLName         xml.Name
	X509Certificate X509Certificate
}

// Transforms contains transformation algorithms applied to the data before hashing in the signature.
type Transforms struct {
	XMLName   xml.Name
	Transform Transform
}

// DigestMethod specifies the algorithm used for digesting the data in the signature.
type DigestMethod struct {
	XMLName   xml.Name
	Algorithm string `xml:"Algorithm,attr"`
}

// DigestValue represents the hashed value of the referenced data in the signature.
type DigestValue struct {
	XMLName xml.Name
}

// X509Certificate represents the X.509 certificate in PEM format.
type X509Certificate struct {
	XMLName xml.Name
	Cert    string `xml:",innerxml"`
}

// Transform represents a single transformation algorithm used in the digital signature process.
type Transform struct {
	XMLName   xml.Name
	Algorithm string `xml:"Algorithm,attr"`
}

// EntityDescriptor represents the root element in SAML metadata, providing information about a SAML entity.
type EntityDescriptor struct {
	XMLName  xml.Name
	DS       string `xml:"xmlns:ds,attr"`
	XMLNS    string `xml:"xmlns,attr"`
	MD       string `xml:"xmlns:md,attr"`
	EntityID string `xml:"entityID,attr"`

	Extensions      Extensions      `xml:"Extensions"`
	SPSSODescriptor SPSSODescriptor `xml:"SPSSODescriptor"`
}

// Extensions represents extensions in the SAML metadata, allowing additional elements.
type Extensions struct {
	XMLName          xml.Name
	Alg              string `xml:"xmlns:alg,attr"`
	MDAttr           string `xml:"xmlns:mdattr,attr"`
	MDRPI            string `xml:"xmlns:mdrpi,attr"`
	EntityAttributes string `xml:"EntityAttributes,omitempty"`
	UIInfo           UIInfo
}

// UIInfo represents user interface information for an entity, such as display name and description.
type UIInfo struct {
	XMLName     xml.Name
	DisplayName UIDisplayName
	MDUI        string `xml:"xmlns:mdui,attr"`
	Description UIDescription
}

// UIDisplayName represents a display name with language information for a SAML entity.
type UIDisplayName struct {
	XMLName xml.Name `xml:"mdui:DisplayName"`
	Lang    string   `xml:"xml:lang,attr,omitempty"`
	Value   string   `xml:",innerxml"`
}

// UIDescription provides a description with language information for a SAML entity.
type UIDescription struct {
	XMLName xml.Name `xml:"mdui:Description"`
	Lang    string   `xml:"xml:lang,attr,omitempty"`
	Value   string   `xml:",innerxml"`
}

// SPSSODescriptor represents the service provider's metadata in SAML.
type SPSSODescriptor struct {
	XMLName                    xml.Name
	ProtocolSupportEnumeration string `xml:"protocolSupportEnumeration,attr"`
	AuthnRequestsSigned        string `xml:"AuthnRequestsSigned,attr"`
	WantAssertionsSigned       string `xml:"WantAssertionsSigned,attr"`
	SigningKeyDescriptor       KeyDescriptor
	EncryptionKeyDescriptor    KeyDescriptor
	AssertionConsumerServices  []AssertionConsumerService
	Extensions                 Extensions `xml:"Extensions"`
}

// EntityAttributes represents attributes associated with a SAML entity.
type EntityAttributes struct {
	XMLName xml.Name
	SAML    string `xml:"xmlns:saml,attr"`

	EntityAttributes []Attribute `xml:"Attribute"` // should be array??
}

// KeyDescriptor represents a key descriptor element, typically containing signing and encryption keys.
type KeyDescriptor struct {
	XMLName xml.Name
	Use     string  `xml:"use,attr"`
	KeyInfo KeyInfo `xml:"KeyInfo"`
}

// AssertionConsumerService represents an endpoint where SAML assertions are sent after authentication.
type AssertionConsumerService struct {
	XMLName  xml.Name
	Binding  string `xml:"Binding,attr"`
	Location string `xml:"Location,attr"`
	Index    string `xml:"index,attr"`
	Default  bool   `xml:"isDefault,attr,omitempty"`
}

// SAMLResponse represents a SAML response message, including assertions and status.
type SAMLResponse struct {
	XMLName      xml.Name
	SAMLP        string `xml:"xmlns:saml2p,attr"`
	SAML         string `xml:"xmlns:saml,attr"`
	SAMLSIG      string `xml:"xmlns:samlsig,attr"`
	Destination  string `xml:"Destination,attr"`
	ID           string `xml:"ID,attr"`
	Version      string `xml:"Version,attr"`
	IssueInstant string `xml:"IssueInstant,attr"`
	InResponseTo string `xml:"InResponseTo,attr"`

	EncryptedAssertion EncryptedAssertion `xml:"EncryptedAssertion,omitempty"`
	Assertion          Assertion          `xml:"Assertion,omitempty"`
	Issuer             Issuer             `xml:"Issuer"`
	Status             Status             `xml:"Status"`
	Signature          Signature          `xml:"Signature"`
	originalString     string
}

// EncryptedAssertion represents an encrypted SAML assertion.
type EncryptedAssertion struct {
	XMLName       xml.Name
	EncryptedData EncryptedData
	Assertion     Assertion `xml:"Assertion"`
}

// EncryptedData represents an encrypted assertion or other encrypted data.
type EncryptedData struct {
	XMLName          xml.Name
	EncryptionMethod EncryptionMethod
	KeyInfo          KeyInfoMain `xml:"KeyInfo"`
	CipherData       CipherData
}

// EncryptionMethod specifies the encryption algorithm used for encryption.
type EncryptionMethod struct {
	XMLName      xml.Name
	Algorithm    string `xml:"Algorithm,attr"`
	DigestMethod DigestMethod
}

// EncryptedKey holds an encrypted key, typically used to encrypt other data.
type EncryptedKey struct {
	XMLName          xml.Name
	ID               string `xml:"Id,attr"`
	EncryptionMethod EncryptionMethod
	KeyInfo          KeyInfo
	CipherData       CipherData
}

// CipherData contains the actual encrypted data or key.
type CipherData struct {
	XMLName     xml.Name
	CipherValue string `xml:"CipherValue"`
}

// Assertion represents a SAML assertion, including conditions and attributes.
type Assertion struct {
	XMLName            xml.Name
	ID                 string `xml:"ID,attr"`
	Version            string `xml:"Version,attr"`
	XS                 string `xml:"xmlns:xs,attr"`
	XSI                string `xml:"xmlns:xsi,attr"`
	SAML               string `xml:"saml,attr"`
	IssueInstant       string `xml:"IssueInstant,attr"`
	Issuer             Issuer `xml:"Issuer"`
	Subject            Subject
	Conditions         Conditions
	AttributeStatement AttributeStatement
	Signature          Signature
}

// Conditions defines constraints for when an assertion is valid.
type Conditions struct {
	XMLName      xml.Name
	NotBefore    string `xml:",attr"`
	NotOnOrAfter string `xml:",attr"`
}

// Subject represents the entity described by the assertion.
type Subject struct {
	XMLName             xml.Name
	NameID              NameID
	SubjectConfirmation SubjectConfirmation
}

// SubjectConfirmation contains methods and data for confirming the subject.
type SubjectConfirmation struct {
	XMLName                 xml.Name
	Method                  string `xml:",attr"`
	SubjectConfirmationData SubjectConfirmationData
}

// Status represents the response status of a SAML operation.
type Status struct {
	XMLName    xml.Name
	StatusCode StatusCode `xml:"StatusCode"`
}

// SubjectConfirmationData holds data for subject confirmation, such as response binding info.
type SubjectConfirmationData struct {
	InResponseTo string `xml:",attr"`
	NotOnOrAfter string `xml:",attr"`
	Recipient    string `xml:",attr"`
}

// NameID represents a subject identifier, such as an email or username.
type NameID struct {
	XMLName xml.Name
	Format  string `xml:",attr"`
	Value   string `xml:",innerxml"`
}

// StatusCode holds the status code of a SAML response.
type StatusCode struct {
	XMLName xml.Name
	Value   string `xml:",attr"`
}

// AttributeValue holds the value of a SAML attribute.
type AttributeValue struct {
	XMLName xml.Name
	Type    string `xml:"xsi:type,attr"`
	Value   string `xml:",innerxml"`
}

// Attribute represents a SAML attribute and its value(s).
type Attribute struct {
	XMLName        xml.Name
	Name           string `xml:",attr"`
	FriendlyName   string `xml:",attr"`
	NameFormat     string `xml:",attr"`
	AttributeValue AttributeValue
}

// AttributeStatement holds a collection of attributes about a subject.
type AttributeStatement struct {
	XMLName    xml.Name
	Attributes []Attribute `xml:"Attribute"`
}

// ParseEncodedResponse decodes a base64-encoded XML SAML response and parses it.
func ParseEncodedResponse(b64ResponseXML string) (*SAMLResponse, error) {
	response := SAMLResponse{}
	bytesXML, err := base64.StdEncoding.DecodeString(b64ResponseXML)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(bytesXML, &response)
	if err != nil {
		return nil, err
	}

	response.originalString = string(bytesXML)
	return &response, nil
}
