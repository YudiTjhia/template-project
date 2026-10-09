import 'package:geolocator/geolocator.dart';

/// Requests location access only when a feature needs the device's location.
class LocationService {
  LocationService._();

  static final instance = LocationService._();

  Future<LocationPermission> checkPermission() => Geolocator.checkPermission();

  Future<LocationPermission> requestPermission() async {
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    return permission;
  }

  Future<bool> isLocationEnabled() => Geolocator.isLocationServiceEnabled();

  /// Requests permission and returns the current position when available.
  /// Returns null when the user denies access or turns off location services.
  Future<Position?> getCurrentPosition() async {
    final enabled = await isLocationEnabled();
    if (!enabled) return null;

    final permission = await requestPermission();
    if (permission == LocationPermission.denied ||
        permission == LocationPermission.deniedForever) {
      return null;
    }

    return Geolocator.getCurrentPosition();
  }

  Future<bool> openAppSettings() => Geolocator.openAppSettings();

  Future<bool> openLocationSettings() => Geolocator.openLocationSettings();
}
