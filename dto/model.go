package dto

type PostVerifyRoamingCert struct {
    FpCode  string `json:"pCode" validate:"required,min=1,max=45"`
    FuserId string `json:"userId" validate:"required,min=1,max=20"`
    ForgId  string `json:"orgId" validate:"required,min=1,max=45"`
}

type PostVerifyRoamingPin struct {
    FpCode  string `json:"pCode" validate:"required,min=1,max=45"`
    FuserId string `json:"userId" validate:"required,min=1,max=20"`
    ForgId  string `json:"orgId" validate:"required,min=1,max=45"`
    Fpin    string `json:"pin" validate:"required,min=1,max=16"`
}

type PostSignPdfConfigB struct {
    FpCode               string `json:"pCode" validate:"required,min=1,max=45"`
    Fsource              string `json:"source" validate:"required,min=1,max=300"`
    Fdest                string `json:"dest" validate:"required,min=1,max=300"`
    FsignerXyPage        string `json:"signerXyPage" validate:"required,min=1,max=300"`
    FsignerImagePath     string `json:"signerImagePath" validate:"max=300"`
    FqrValue             string `json:"qrValue" validate:"max=200"`
    FqrPosition          string `json:"qrPosition" validate:"max=300"`
    FqrSize              int    `json:"qrSize" validate:"max=3" default:"0"`
    FsignerFontSize      int    `json:"signerFontSize" validate:"max=3"`
    FcustomSignatureText string `json:"customSignatureText" validate:"max=80"`
    FsignerTextRow       int    `json:"signerTextRow" validate:"max=3"`
    FdtsFlag             int    `json:"dtsFlag" validate:"required,min=1,max=1" default:"1"`
    FuserId              string `json:"userId" validate:"required,min=1,max=20"`
    ForgId               string `json:"orgId" validate:"required,min=1,max=45"`
    Fpin                 string `json:"pin" validate:"required,min=1,max=16"`
    Fremark              string `json:"remark" validate:"max=300"`
    FfileServerId        string `json:"fileServerId" validate:"required,min=1,max=20"`
}
