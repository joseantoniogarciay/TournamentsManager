import { Redirect } from "expo-router";

// Una URL externa malformada nunca expone la pantalla interna de Expo Router.
export default function UnmatchedRoute() {
  return <Redirect href="/" />;
}
