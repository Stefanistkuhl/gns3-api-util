from . import GNS3APIClient
import urllib.parse


class GNS3PostAPI(GNS3APIClient):

    # controller endpoints

    def check_version(self, data):
        """Check if version is the same as the server."""
        return self._api_call("version", method="POST", data=data)

    def shutdown(self):
        """Shutdown the controller"""
        return self._api_call("shutdown", method="POST")

    def login(self):
        """Login as user"""
        return self._api_call("shutdown", method="POST")

    # user endpoints
    # todo add the authenticatie endpoint i am too lazy for rn

    def user(self, data):
        """Check if version is the same as the server."""
        return self._api_call("access/users", method="POST", data=data)
