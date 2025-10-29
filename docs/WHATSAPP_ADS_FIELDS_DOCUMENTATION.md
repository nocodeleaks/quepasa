# WhatsApp Ads Fields Documentation

## Overview

This document describes all fields available in WhatsApp advertisement messages (Click-to-WhatsApp Ads) received through the WhatsApp Business API. These fields are extracted from `ExternalAdReply` and `ContextInfo` structures in the WhatsApp protobuf messages.

---

## Table of Contents

1. [ExternalAdReply Fields](#externaladreply-fields)
2. [ContextInfo Conversion Fields](#contextinfo-conversion-fields)
3. [Usage Examples](#usage-examples)
4. [Integration Notes](#integration-notes)

---

## ExternalAdReply Fields

These fields come from the `ExternalAdReply` structure within the message's `ContextInfo`, providing details about the advertisement that triggered the conversation.

### Core Identification

#### `id` (string)
- **Description**: Click-to-WhatsApp ID (CTWA Click ID)
- **Purpose**: Unique identifier for tracking the specific ad click that initiated the conversation
- **Format**: Base64-encoded string
- **Example**: `"Afdrf5n4Pylg1rQCwk43dvvNTKLF-KSXCIvEm9IGFW29ER51drkrGyx9KTT2T5hn..."`
- **Use Case**: Analytics, attribution tracking, ad performance measurement

#### `title` (string)
- **Description**: Ad headline or call-to-action text
- **Purpose**: The main text displayed in the ad button
- **Example**: `"Converse conosco"`, `"Garanta seu Desconto na Blue Friday"`
- **Use Case**: Identifying which ad creative generated the lead

#### `sourceid` (string)
- **Description**: Facebook/Meta ad or post identifier
- **Purpose**: Links the message back to the specific ad/post in Meta's advertising system
- **Format**: Numeric string (120+ digits)
- **Example**: `"120234805578560256"`
- **Use Case**: Cross-referencing with Facebook Ads Manager, ROI tracking

#### `sourceurl` (string)
- **Description**: Deep link or tracking URL for the ad
- **Purpose**: URL that was associated with the ad when created
- **Format**: Shortened Facebook URL
- **Example**: `"https://fb.me/8wiTABFVY"`
- **Use Case**: Tracking user journey, UTM parameter analysis

### Ad Content

#### `body` (string)
- **Description**: Ad description or body text
- **Purpose**: The detailed text content shown in the ad preview
- **Example**: `"💄 Lábios mais definidos e hidratados com resultado imediato!..."`
- **Use Case**: Understanding what messaging resonated with the user

#### `mediatype` (string)
- **Description**: Type of media in the ad
- **Possible Values**: `"IMAGE"`, `"VIDEO"`, `"NONE"`
- **Purpose**: Indicates the format of the ad creative
- **Use Case**: Analyzing which media types drive better engagement

#### `thumbnailurl` (string)
- **Description**: URL to the ad's thumbnail/preview image
- **Purpose**: Link to the visual content shown in the ad
- **Format**: Facebook CDN URL with query parameters
- **Example**: `"https://scontent.xx.fbcdn.net/v/t45.1600-4/565742744_2099243270612575..."`
- **Use Case**: Displaying ad preview in CRM/support systems

#### `originalimageurl` (string)
- **Description**: URL to the full-resolution original ad image
- **Purpose**: Link to high-quality version of the ad creative
- **Format**: Facebook Ads image URL
- **Example**: `"https://www.facebook.com/ads/image/?d=AQJb3BIvhBvEZDGM5egaEmc8D626Ydjaxi0S8sxzzT2x..."`
- **Use Case**: Quality assurance, creative analysis

### Source Attribution

#### `sourcetype` (string)
- **Description**: Type of Facebook content that triggered the message
- **Possible Values**: `"ad"` (paid advertisement), `"post"` (organic post with CTA)
- **Purpose**: Distinguishes between paid ads and organic posts
- **Use Case**: Budget allocation, organic vs paid performance analysis

#### `app` (string)
- **Description**: Meta platform where the ad was shown
- **Possible Values**: `"facebook"`, `"instagram"`, `"messenger"`
- **Purpose**: Identifies which platform drove the conversation
- **Use Case**: Multi-platform campaign attribution

#### `type` (string)
- **Description**: Duplicate of `sourcetype` (legacy field)
- **Purpose**: Backwards compatibility
- **Note**: Same as `sourcetype`, may be deprecated in future

### User Experience Flags

#### `containsautoreply` (boolean)
- **Description**: Indicates if the ad has automated reply configured
- **Purpose**: Whether an automatic greeting was sent to the user
- **Default**: `false`
- **Use Case**: Automation tracking, greeting message management

#### `renderlargerthumbnail` (boolean)
- **Description**: Display preference for thumbnail size
- **Purpose**: Whether to render the ad preview in larger format
- **Default**: Usually `true` for Click-to-WhatsApp ads
- **Use Case**: UI rendering decisions

#### `showadattribution` (boolean)
- **Description**: Whether to display "Sponsored" label or ad attribution
- **Purpose**: Compliance with advertising disclosure requirements
- **Default**: Usually `true` for paid ads
- **Use Case**: Transparency, regulatory compliance

#### `clicktowhatsappcall` (boolean)
- **Description**: Indicates if this is a Click-to-WhatsApp Call ad
- **Purpose**: Whether the ad specifically promoted calling via WhatsApp
- **Default**: `false` for chat ads, `true` for call ads
- **Use Case**: Call tracking, analyzing call vs chat preference

#### `adcontextpreviewdismissed` (boolean)
- **Description**: User action flag for ad preview dismissal
- **Purpose**: Whether the user dismissed the ad context/preview
- **Default**: `false`
- **Use Case**: User engagement analysis

### Automated Messaging

#### `automatedgreetingmessageshown` (boolean)
- **Description**: Status of automated greeting delivery
- **Purpose**: Confirms if an auto-greeting was sent to the user
- **Default**: `false`
- **Use Case**: Automation verification, user experience tracking

#### `greetingmessagebody` (string)
- **Description**: Content of the automated greeting message
- **Purpose**: The actual text sent as automatic first message
- **Example**: `"Oi! Como podemos ajudar?"`
- **Use Case**: Greeting personalization, A/B testing greetings

#### `disablenudge` (boolean)
- **Description**: Nudge notification control flag
- **Purpose**: Whether follow-up nudges are disabled for this conversation
- **Default**: `false`
- **Use Case**: User experience control, preventing notification spam

---

## ContextInfo Conversion Fields

These fields come from the `ContextInfo` structure and provide detailed attribution and conversion tracking data.

### Conversion Tracking

#### `conversionsource` (string)
- **Description**: Primary conversion source identifier
- **Possible Values**: 
  - `"FB_Ads"` - Facebook Ads
  - `"FB_Post"` - Facebook Post with CTA
  - `"IG_Ads"` - Instagram Ads
  - `"ctwa_ad"` - Click-to-WhatsApp Ad (legacy)
- **Purpose**: Identifies the Meta product that drove the conversion
- **Use Case**: Source attribution, ROI calculation per platform

#### `conversiondata` (string)
- **Description**: Encrypted conversion attribution payload
- **Purpose**: Secure attribution data for Facebook Conversion API
- **Format**: Base64-encoded encrypted string (200-400 characters)
- **Example**: `"AfeNj133t3MIy-cGr5gkiEOBWjywFNSBu3yuit0Asac91KqKJUo4OmuVayIOjt4ePFu..."`
- **Use Case**: 
  - Facebook Conversion API integration
  - Server-side event tracking
  - Advanced attribution models
  - Privacy-preserving measurement

#### `conversiondelayseconds` (uint32, nullable)
- **Description**: Time delay between ad click and message send
- **Purpose**: Measures user hesitation/consideration time
- **Unit**: Seconds
- **Typical Values**: 0-60 seconds (most users message within 1 minute)
- **Example**: `10` (user waited 10 seconds after clicking ad)
- **Use Case**: 
  - User behavior analysis
  - Ad creative effectiveness
  - Conversion velocity tracking
- **Note**: This field changes on message retries (can go from 10→3→0) but doesn't affect message content

### Entry Point Attribution

#### `entrypointconversionsource` (string)
- **Description**: Specific entry point that triggered the conversion
- **Possible Values**:
  - `"ctwa_ad"` - Click-to-WhatsApp advertisement
  - `"post_cta"` - Post call-to-action button
  - `"page_cta"` - Page CTA button
  - `"story_mention"` - Instagram story mention
- **Purpose**: Granular attribution of the exact UI element clicked
- **Use Case**: A/B testing different ad formats, optimizing placement

#### `entrypointconversionapp` (string)
- **Description**: Meta app where the entry point was located
- **Possible Values**: `"facebook"`, `"instagram"`, `"messenger"`
- **Purpose**: Platform-specific attribution
- **Use Case**: Cross-platform campaign performance comparison

#### `entrypointconversiondelayseconds` (uint32, nullable)
- **Description**: Delay specific to the entry point interaction
- **Purpose**: Measures time from entry point click to conversion
- **Unit**: Seconds
- **Typical Values**: Usually same as or slightly less than `conversiondelayseconds`
- **Example**: `9` (1 second faster than overall conversion delay)
- **Use Case**: Entry point UX optimization

---

## Usage Examples

### Example 1: Facebook Click-to-WhatsApp Ad

```json
{
  "ads": {
    "id": "Afdrf5n4Pylg1rQCwk43dvvNTKLF-KSXCIvEm9IGFW29ER51drkrGyx9KTT2T5hn...",
    "title": "Converse conosco",
    "body": "💄 Lábios mais definidos e hidratados com resultado imediato!...",
    "mediatype": "IMAGE",
    "thumbnailurl": "https://scontent.xx.fbcdn.net/v/t45.1600-4/...",
    "sourceid": "120234805578560256",
    "sourceurl": "https://fb.me/8wiTABFVY",
    "sourcetype": "ad",
    "app": "facebook",
    "conversionsource": "FB_Ads",
    "conversiondelayseconds": 10,
    "entrypointconversionsource": "ctwa_ad",
    "entrypointconversionapp": "facebook",
    "entrypointconversiondelayseconds": 9,
    "automatedgreetingmessageshown": true,
    "greetingmessagebody": "Oi! Como podemos ajudar?"
  }
}
```

**Interpretation**: User saw a Facebook Click-to-WhatsApp ad with an image about lip fillers, waited 10 seconds after clicking before sending a message, and received an automated greeting.

### Example 2: Instagram Post with CTA

```json
{
  "ads": {
    "id": "ARACGmmFDy48IyrcMV2igjacnY39gcMjcUUlc_WBbfSBaoJU5Kxb7WjxI7j03Dx8...",
    "title": "Garanta seu Desconto na Blue Friday",
    "body": "",
    "mediatype": "IMAGE",
    "sourcetype": "post",
    "app": "facebook",
    "conversionsource": "FB_Post",
    "conversiondelayseconds": 3,
    "entrypointconversionsource": "post_cta",
    "entrypointconversionapp": "facebook",
    "showadattribution": false
  }
}
```

**Interpretation**: User clicked a call-to-action on an organic Facebook post (not a paid ad) promoting a Black Friday discount, messaged within 3 seconds (high intent), no automated greeting configured.

---

## Integration Notes

### For CRM/Support Systems

1. **Lead Source Tracking**: Use `conversionsource`, `sourcetype`, and `app` fields to categorize leads
2. **Ad Performance**: Track `sourceid` to measure which specific ads generate the most quality leads
3. **User Intent**: Analyze `conversiondelayseconds` - lower values often indicate higher purchase intent
4. **Greeting Management**: Check `automatedgreetingmessageshown` to avoid duplicate greetings

### For Analytics Platforms

1. **Attribution**: Send `conversiondata` to Facebook Conversion API for accurate attribution
2. **Campaign ROI**: Join `sourceid` with Facebook Ads Manager data
3. **Platform Comparison**: Group by `app` to compare Facebook vs Instagram performance
4. **Content Analysis**: Track which `title` and `body` combinations drive best results

### For Automation Systems

1. **Conditional Flows**: Skip initial greeting if `automatedgreetingmessageshown` is `true`
2. **Personalization**: Use `title` and `body` to personalize follow-up messages
3. **Priority Routing**: Fast conversions (`conversiondelayseconds` < 5) may indicate high-intent users

### Important Considerations

#### Field Stability
- **Stable Fields**: `id`, `title`, `body`, `sourceid`, `sourceurl`, `sourcetype`, `app`, `conversionsource`
- **Volatile Fields**: `conversiondelayseconds`, `entrypointconversiondelayseconds` (change on retries)
- **Recommendation**: Do NOT use delay fields for duplicate detection or caching logic

#### Privacy & Compliance
- `conversiondata` contains encrypted user attribution data - handle according to privacy policies
- `showadattribution` ensures regulatory compliance with advertising disclosure
- Do not expose raw `conversiondata` to end users

#### Null/Empty Values
- Boolean fields default to `false` if not present
- String fields may be empty (`""`) vs null - check both conditions
- `conversiondelayseconds` and `entrypointconversiondelayseconds` are nullable pointers

---

## Field Availability Matrix

| Field | Click-to-WhatsApp Ad | Organic Post CTA | Always Present |
|-------|---------------------|------------------|----------------|
| `id` | ✅ | ✅ | ✅ |
| `title` | ✅ | ✅ | ✅ |
| `body` | ✅ | ⚠️ (often empty) | ❌ |
| `mediatype` | ✅ | ✅ | ✅ |
| `sourceid` | ✅ | ✅ | ✅ |
| `sourceurl` | ✅ | ✅ | ✅ |
| `sourcetype` | ✅ (`"ad"`) | ✅ (`"post"`) | ✅ |
| `app` | ✅ | ✅ | ✅ |
| `conversionsource` | ✅ | ✅ | ✅ |
| `conversiondata` | ✅ | ✅ | ✅ |
| `conversiondelayseconds` | ✅ | ✅ | ⚠️ (may be null) |
| `entrypointconversionsource` | ✅ | ✅ | ✅ |
| `automatedgreetingmessageshown` | ⚠️ (if configured) | ⚠️ (if configured) | ❌ |
| `greetingmessagebody` | ⚠️ (if configured) | ⚠️ (if configured) | ❌ |

**Legend**:
- ✅ Always present with meaningful value
- ⚠️ Conditionally present
- ❌ May be absent or empty

---

## References

- [WhatsApp Business Platform Documentation](https://developers.facebook.com/docs/whatsapp/)
- [Click-to-WhatsApp Ads](https://www.facebook.com/business/ads/click-to-whatsapp-ads)
- [Facebook Conversion API](https://developers.facebook.com/docs/marketing-api/conversions-api/)
- [WhatsApp Cloud API - Messages](https://developers.facebook.com/docs/whatsapp/cloud-api/guides/send-messages)

---

## Version History

- **v1.0** (2025-10-29): Initial documentation covering all 19 ExternalAdReply fields and 6 ContextInfo conversion fields
- Based on real production logs from Click-to-WhatsApp ad campaigns

---

## Support

For questions or issues related to these fields:
1. Check the [WhatsApp Business Platform Support](https://developers.facebook.com/docs/whatsapp/support)
2. Review the [Facebook Ads Help Center](https://www.facebook.com/business/help/)
3. Consult the [whatsmeow library documentation](https://pkg.go.dev/go.mau.fi/whatsmeow)
