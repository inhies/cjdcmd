using System;
using System.Linq;
using MyCompany;
using System.Web;
using System.Web.Security;
using MyCompany.DAL;
using MyCompany.Globalization;
using MyCompany.DAL.Logs;
using MyCompany.Logging;

namespace MyCompany
{

    public class Auth
    {

        public class AuthException : Exception
        {
            public int StatusCode = 0;
            public AuthException(string message, int statusCode) : base(message) { StatusCode = statusCode;  }
        }

        public class EmptyEmailException : AuthException
        {
            public EmptyEmailException() : base(Language.RES_ERROR_LOGIN_CLIENT_EMPTY_EMAIL, 6) { }
        }

        public class EmptyPasswordException : AuthException
        {
            public EmptyPasswordException() : base(Language.RES_ERROR_LOGIN_CLIENT_EMPTY_PASSWORD, 7) { }
        }

        public class WrongEmailException : AuthException
        {
            public WrongEmailException() : base(Language.RES_ERROR_LOGIN_CLIENT_WRONG_EMAIL, 2) { }
        }

        public class WrongPasswordException : AuthException
        {
            public WrongPasswordException() : base(Language.RES_ERROR_LOGIN_CLIENT_WRONG_PASSWORD, 3) { }
        }

        public class InactiveAccountException : AuthException
        {
            public InactiveAccountException() : base(Language.RES_ERROR_LOGIN_CLIENT_INACTIVE_ACCOUNT, 5) { }
        }

        public class EmailNotValidatedException : AuthException
        {
            public EmailNotValidatedException() : base(Language.RES_ERROR_LOGIN_CLIENT_EMAIL_NOT_VALIDATED, 4) { }
        }

        private readonly string CLIENT_KEY = "9A751E0D-816F-4A92-9185-559D38661F77";

        private readonly string CLIENT_USER_KEY = "0CE2F700-1375-4B0F-8400-06A01CED2658";

        public Client Client
        {
            get
            {
                if(!IsAuthenticated) return null;
                if(HttpContext.Current.Items[CLIENT_KEY]==null)
                {
                    HttpContext.Current.Items[CLIENT_KEY] = ClientMethods.Get<Client>((Guid)ClientId); 
                }
                return (Client)HttpContext.Current.Items[CLIENT_KEY];
            }
        }

        public ClientUser ClientUser
        {
            get
            {
                if (!IsAuthenticated) return null;
                if (HttpContext.Current.Items[CLIENT_USER_KEY] == null)
                {
                    HttpContext.Current.Items[CLIENT_USER_KEY] = ClientUserMethods.GetByClientId((Guid)ClientId);
                }
                return (ClientUser)HttpContext.Current.Items[CLIENT_USER_KEY];
            }
        }

        public Boolean IsAuthenticated { get; set; }

        public Guid? ClientId { 
            get 
            {
                if (!IsAuthenticated) return null;
                return (Guid)HttpContext.Current.Session["ClientId"];
            } 
        }

        public Guid? ClientUserId { 
            get {
                if (!IsAuthenticated) return null;
                return ClientUser.Id;
            } 
        }

        public int ClientTypeId { 
            get {
                if (!IsAuthenticated) return 0;
                return Client.ClientTypeId;
            } 
        }

        public Auth()
        {
            if (HttpContext.Current.User.Identity.IsAuthenticated)
            {
                IsAuthenticated = true;
            }
        }

        public void RequireClientOfType(params int[] types)
        {
            if (!(IsAuthenticated && types.Contains(ClientTypeId)))
            {
                HttpContext.Current.Response.Redirect((new UrlFactory(false)).GetHomeUrl(), true);
            }
        }

        public void Logout()
        {
            Logout(true);
        }

        public void Logout(Boolean redirect)
        {
            FormsAuthentication.SignOut();
            IsAuthenticated = false;
            HttpContext.Current.Session["ClientId"] = null;
            HttpContext.Current.Items[CLIENT_KEY] = null;
            HttpContext.Current.Items[CLIENT_USER_KEY] = null;
            if(redirect) HttpContext.Current.Response.Redirect((new UrlFactory(false)).GetHomeUrl(), true);
        }

        public void Login(string email, string password, bool autoLogin)
        {
            Logout(false);

            email = email.Trim().ToLower();
            password = password.Trim();

            int status = 1;

            LoginAttemptLog log = new LoginAttemptLog { AutoLogin = autoLogin, Email = email, Password = password };

            try
            {
                if (string.IsNullOrEmpty(email)) throw new EmptyEmailException();

                if (string.IsNullOrEmpty(password)) throw new EmptyPasswordException();

                ClientUser clientUser = ClientUserMethods.GetByEmailExcludingProspects(email);

                if (clientUser == null) throw new WrongEmailException();

                if (!clientUser.Password.Equals(password)) throw new WrongPasswordException();

                Client client = clientUser.Client;

                if (!(bool)client.PreRegCheck) throw new EmailNotValidatedException();

                if (!(bool)client.Active || client.DeleteFlag.Equals("y")) throw new InactiveAccountException();

                FormsAuthentication.SetAuthCookie(client.Id.ToString(), true);
                HttpContext.Current.Session["ClientId"] = client.Id;

                log.KeyId = client.Id;
                log.KeyEntityId = ClientMethods.GetEntityId(client.ClientTypeId);
            }
            catch (AuthException ax)
            {
                status = ax.StatusCode;
                log.Success = status == 1;
                log.Status = status;
            }
            finally
            {
                LogRecorder.Record(log);
            }

        }

    }

}
