package kbt;

import java.util.ArrayList;
import java.util.List;

import org.jboss.logging.Logger;
import org.json.JSONObject;
import org.keycloak.Config;
import org.keycloak.authentication.FormContext;
import org.keycloak.authentication.ValidationContext;
import org.keycloak.authentication.forms.RegistrationUserCreation;
import org.keycloak.forms.login.LoginFormsProvider;
import org.keycloak.http.HttpRequest;
import org.keycloak.models.GroupModel;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.UserModel;
import org.keycloak.models.utils.FormMessage;
import org.keycloak.representations.idm.AbstractUserRepresentation;
import org.keycloak.userprofile.UserProfile;
import org.keycloak.utils.RegexUtils;

import jakarta.ws.rs.core.MultivaluedHashMap;
import jakarta.ws.rs.core.MultivaluedMap;
import jakarta.ws.rs.ext.Provider;

@Provider
public class CustomUserRegistration extends RegistrationUserCreation {

  private static final Logger log = Logger.getLogger(CustomEventListenerProvider.class);

  @Override
  public void init(Config.Scope config) {
  }

  @Override
  public String getId() {
    return "custom-registration-user-creation";
  }

  @Override
  public String getDisplayType() {
    return "Custom Registration User Creation";
  }

  @Override
  public void buildPage(FormContext context, LoginFormsProvider form) {
    KeycloakSession session = context.getSession();

    HttpRequest request = context.getHttpRequest();
    MultivaluedMap<String, String> queryParams = request.getUri().getQueryParameters();

    String groupCode = queryParams.getFirst("groupCode");
    Boolean failedValidation = groupCode == null;

    if (failedValidation) {
      try {
        groupCode = request != null ? request.getDecodedFormParameters().getFirst("groupCode") : null;
      } catch (Exception e) {
      }
    }

    Boolean suppliedCode = groupCode != null && groupCode.length() > 0;

    if (suppliedCode) {
      String groupName = (String) session.getAttribute("groupName");
      String allowedDomains = (String) session.getAttribute("allowedDomains");

      if (groupName == null || allowedDomains == null || groupCode != (String) session.getAttribute("groupCode")) {
        // This block is getting basic group code info
        JSONObject registrationValidationPayload = new JSONObject();
        registrationValidationPayload.put("groupCode", groupCode);
        JSONObject registrationValidationResponse = BackchannelAuth.sendUnixMessage("REGISTER_VALIDATE",
            registrationValidationPayload);

        if (registrationValidationResponse.getBoolean("success")) {
          String groupId = registrationValidationResponse.getString("id");
          groupName = registrationValidationResponse.getString("name").replaceAll("_", " ");
          allowedDomains = registrationValidationResponse.getString("allowedDomains").replaceAll(",", ", ");
          session.setAttribute("groupId", groupId);
          session.setAttribute("groupCode", groupCode);
          session.setAttribute("groupName", groupName);
          session.setAttribute("allowedDomains", allowedDomains);
        } else if (registrationValidationResponse.getString("reason").contains("BAD_GROUP")) {
          form.setErrors(List.of(new FormMessage("groupCode", "invalidGroup")));
        }
      }

      if (!failedValidation) {
        MultivaluedMap<String, String> formData = new MultivaluedHashMap<String, String>();
        formData.putSingle("groupCode", groupCode);
        form.setFormData(formData);
      }

      form
          .setAttribute("groupName", groupName)
          .setAttribute("allowedDomains", allowedDomains);

    } else {
      session.removeAttribute("groupCode");
      session.removeAttribute("groupName");
      session.removeAttribute("allowedDomains");
    }

  }

  @Override
  public void validate(ValidationContext context) {
    KeycloakSession session = context.getSession();

    List<FormMessage> validationErrors = new ArrayList<>();
    MultivaluedMap<String, String> formData = context.getHttpRequest().getDecodedFormParameters();

    if (formData.containsKey("groupCode")) {
      String groupCode = formData.getFirst("groupCode");
      String email = formData.getFirst("email");

      String groupName = (String) session.getAttribute("groupName");
      String allowedDomains = (String) session.getAttribute("allowedDomains");

      Boolean suppliedCode = groupCode != null && groupCode.length() > 0;

      if ((groupName == null || allowedDomains == null) && suppliedCode) {

        if (!RegexUtils.valueMatchesRegex("[a-zA-Z0-9]{8}", groupCode)) {
          validationErrors.add(new FormMessage("groupCode", "invalidGroup"));
        } else {

          if (groupCode != session.getAttribute("groupCode")) {

            // Get group information for registration
            JSONObject registrationValidationPayload = new JSONObject();
            registrationValidationPayload.put("groupCode", groupCode);
            JSONObject registrationValidationResponse = BackchannelAuth.sendUnixMessage("REGISTER_VALIDATE",
                registrationValidationPayload);

            if (false == registrationValidationResponse.getBoolean("success")) {

              String reason = registrationValidationResponse.getString("reason");

              if (reason.contains("BAD_GROUP")) {
                validationErrors.add(new FormMessage("groupCode", "invalidGroup"));
              }
            } else {
              String groupId = registrationValidationResponse.getString("id");
              groupName = registrationValidationResponse.getString("name");
              allowedDomains = registrationValidationResponse.getString("allowedDomains");
              session.setAttribute("groupId", groupId);
              session.setAttribute("groupCode", groupCode);
              session.setAttribute("groupName", groupName);
              session.setAttribute("allowedDomains", allowedDomains);
            }
          }
        }
      }

      if (allowedDomains != null) {
        List<String> domains = List.of(allowedDomains.split(","));
        if (email == null
            || email.contains("@") && allowedDomains.length() > 0 && !domains.contains(email.split("@")[1])) {
          validationErrors.add(new FormMessage("email", "invalidEmail"));
        }
      } else if (suppliedCode) {
        validationErrors.add(new FormMessage("groupCode", "invalidGroup"));
      }

      UserModel userWithEmail = context.getSession().users().getUserByEmail(context.getRealm(), email);

      if (userWithEmail != null) {
        validationErrors.add(new FormMessage("email", "emailInUse"));
      }

      if (!validationErrors.isEmpty()) {
        context.error("VALIDATION_ERROR");
        context.validationError(formData, validationErrors);
        return;
      }

      // context.getEvent().detail("group_code", groupCode);
    }

    super.validate(context);
  }

  @Override
  public void success(FormContext context) {
    super.success(context);

    KeycloakSession session = context.getSession();

    UserProfile profile = (UserProfile) session.getAttribute("UP_REGISTER");
    AbstractUserRepresentation up = profile.toRepresentation();

    Object groupIdObj = session.getAttribute("groupId");

    String xForwardedFor = context.getHttpRequest().getHttpHeaders().getHeaderString("X-Forwarded-For").split(":")[0];

    JSONObject registrationSuccessPayload = new JSONObject();

    if (groupIdObj != null) {
      GroupModel gr = context.getRealm().getGroupById(groupIdObj.toString());
      context.getUser().joinGroup(gr);
      registrationSuccessPayload.put("groupCode", session.getAttribute("groupCode").toString());
    }

    registrationSuccessPayload.put("userId", up.getId());
    registrationSuccessPayload.put("firstName", up.getFirstName());
    registrationSuccessPayload.put("lastName", up.getLastName());
    registrationSuccessPayload.put("email", up.getEmail());
    registrationSuccessPayload.put("ipAddress", xForwardedFor);

    BackchannelAuth.sendUnixMessage("REGISTER", registrationSuccessPayload);
  }

}
