package cmd

import (
	"anvil/internal"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func getMinecraftProfile(accessToken string) (internal.MinecraftProfile, error) {
    var result internal.MinecraftProfile

    resp, err := rClient.R().
        SetHeader("Authorization", "Bearer "+accessToken).
        SetResult(&result).
        Get("https://api.minecraftservices.com/minecraft/profile")

    if err != nil || resp.IsError() {
        return result, fmt.Errorf("profile fetch failed: %v", err)
    }
    return result, nil
}

func authenticateMinecraft(xstsToken, uhs string) (internal.MinecraftTokenResponse, error) {
    var result internal.MinecraftTokenResponse

    resp, err := rClient.R().
        SetHeader("Content-Type", "application/json").
        SetBody(map[string]interface{}{
            "identityToken": fmt.Sprintf("XBL3.0 x=%s;%s", uhs, xstsToken),
        }).
        SetResult(&result).
        Post("https://api.minecraftservices.com/authentication/login_with_xbox")

    if err != nil || resp.IsError() {
        fmt.Printf("resp.Body(): %v\n", resp.Body())
        return result, fmt.Errorf("minecraft auth failed: %v", err)
    }
    return result, nil
}

func authenticateXSTS(xblToken string) (internal.XSTSResponse, error) {
    var result internal.XSTSResponse

    resp, err := rClient.R().
        SetHeader("Content-Type", "application/json").
        SetHeader("Accept", "application/json").
        SetBody(map[string]interface{}{
            "Properties": map[string]interface{}{
                "SandboxId":  "RETAIL",
                "UserTokens": []string{xblToken},
            },
            "RelyingParty": "rp://api.minecraftservices.com/",
            "TokenType":    "JWT",
        }).
        SetResult(&result).
        Post("https://xsts.auth.xboxlive.com/xsts/authorize")

    if err != nil || resp.IsError() {
        return result, fmt.Errorf("xsts auth failed: %v", err)
    }
    return result, nil
}

func authenticateXBL(accessToken string) (internal.XBLResponse, error) {
    var result internal.XBLResponse

    resp, err := rClient.R().
        SetHeader("Content-Type", "application/json").
        SetHeader("Accept", "application/json").
        SetBody(map[string]interface{}{
            "Properties": map[string]interface{}{
                "AuthMethod": "RPS",
                "SiteName":   "user.auth.xboxlive.com",
                "RpsTicket":  "d=" + accessToken,
            },
            "RelyingParty": "http://auth.xboxlive.com",
            "TokenType":    "JWT",
        }).
        SetResult(&result).
        Post("https://user.auth.xboxlive.com/user/authenticate")

    if err != nil || resp.IsError() {
        return result, fmt.Errorf("xbl auth failed: %v", err)
    }
    return result, nil
}

func pollToken(deviceCode internal.DeviceCodeResponse) (internal.TokenResponse, error) {
    var result internal.TokenResponse

    for {
        time.Sleep(time.Duration(deviceCode.Interval) * time.Second)

        resp, err := rClient.R().
            SetHeader("Content-Type", "application/x-www-form-urlencoded").
            SetBody(fmt.Sprintf(
                "grant_type=urn:ietf:params:oauth:grant-type:device_code&client_id=%s&device_code=%s",
                internal.LauncherClientID, deviceCode.DeviceCode,
            )).
            SetResult(&result).
            Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token")

        if err != nil {
            return result, err
        }

        if resp.IsError() && result.Error == "authorization_pending" {
            continue
        }

        if result.Error != "" {
            return result, fmt.Errorf("token error: %s", result.Error)
        }

        if result.AccessToken != "" {
            return result, nil
        }
    }
}

func getDeviceCode() (internal.DeviceCodeResponse, error) {
    var result internal.DeviceCodeResponse
    resp, err := rClient.R().
        SetHeader("Content-Type", "application/x-www-form-urlencoded").
        SetBody("client_id="+internal.LauncherClientID+"&scope=XboxLive.signin%20offline_access").
        SetResult(&result).
        Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode")
	
	if err != nil || resp.IsError() {
        fmt.Printf("resp.Body(): %v\n", resp.Body())
		return result, fmt.Errorf("device code request failed: %v", err)
	}
	return result, nil
}

func Login(...string) {
    rClient = internal.NewClient()
	deviceCode, err := getDeviceCode();
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Go to %s and enter code: %s\n", deviceCode.VerificationURI, deviceCode.UserCode)

	token, err := pollToken(deviceCode)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := authenticateXBL(token.AccessToken)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	xstsResp, err := authenticateXSTS(resp.Token)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	uhs := xstsResp.DisplayClaims.Xui[0].Uhs
	mcTokenResp, err := authenticateMinecraft(xstsResp.Token, uhs)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	profile, err := getMinecraftProfile(mcTokenResp.AccessToken)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

    profile.AccessToken = mcTokenResp.AccessToken

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(filepath.Join(internal.AnvilHome, "profile.json"), data, 0644)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Logged in as %s\n", profile.Name)
}
	