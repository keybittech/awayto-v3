package kbt;

import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.OutputStreamWriter;

import org.json.JSONObject;
import org.keycloak.models.ClientModel;
import org.keycloak.models.RealmModel;
import org.newsclub.net.unix.AFUNIXSocket;
import org.newsclub.net.unix.AFUNIXSocketAddress;

public final class BackchannelAuth {

  private BackchannelAuth() {
  }

  private static final File SOCKET_FILE = new File(
      System.getenv("KC_UNIX_SOCK_DIR") + "/" + System.getenv("KC_UNIX_SOCK_LOC"));
  // private static final Logger log =
  // Logger.getLogger(CustomEventListenerProvider.class);

  public static String getClientSecret(RealmModel realm) {
    String clientId = System.getenv("KC_API_CLIENT_ID");
    ClientModel client = realm.getClientByClientId(clientId);
    if (client != null) {
      return client.getSecret();
    } else {
      return null;
    }
  }

  public static JSONObject sendUnixMessage(String eventType, JSONObject eventPayload) {
    JSONObject response = new JSONObject();
    response.put("success", false);

    try (AFUNIXSocket sock = AFUNIXSocket.newInstance()) {

      eventPayload.put("webhookName", eventType);

      sock.connect(AFUNIXSocketAddress.of(SOCKET_FILE));
      try (OutputStream out = sock.getOutputStream();
          BufferedWriter writer = new BufferedWriter(new OutputStreamWriter(out))) {
        writer.write(eventPayload.toString() + "\n");
        writer.flush();
      }
      try (InputStream in = sock.getInputStream();
          BufferedReader reader = new BufferedReader(new InputStreamReader(in))) {
        StringBuilder sb = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
          sb.append(line);
        }

        if (sb.toString().length() > 0) {
          return new JSONObject(sb.toString().trim());
        }
      }

    } catch (Exception e) {
      e.printStackTrace();
    }

    return response;
  }
}
